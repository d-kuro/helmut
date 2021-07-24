package helmut_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/d-kuro/helmut"
	"github.com/d-kuro/helmut/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

const volumes = `
volumes:
  - name: aaa
    hostPath:
      path: /aaa
      type: Directory
  - name: bbb
    hostPath:
      path: /bbb
      type: Directory
`

var hostPathDirectory = corev1.HostPathDirectory

func TestRenderTemplates(t *testing.T) {
	t.Parallel()

	const (
		releaseName = "foo"
		chartName   = "test-chart"
	)

	tests := []struct {
		name          string
		options       []helmut.Option
		assertOptions []assert.Option
		crateValues   []byte
		want          runtime.Object
	}{
		{
			name: "contains service account",
			assertOptions: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
			want: newServiceAccount(chartName, releaseName),
		},
		{
			name: "contains service",
			assertOptions: []assert.Option{
				assert.WithIgnoreLabelKeys(
					"app.kubernetes.io/managed-by",
					"app.kubernetes.io/version",
					"helm.sh/chart",
				),
			},
			want: newService(chartName, releaseName,
				withServiceLabels(map[string]string{
					"app.kubernetes.io/name":     chartName,
					"app.kubernetes.io/instance": releaseName,
				})),
		},
		{
			name: "contains deployment",
			assertOptions: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
			want: newDeployment(chartName, releaseName),
		},
		{
			name:    "set helm values",
			options: []helmut.Option{helmut.WithSet("replicaCount=2")},
			assertOptions: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
			want: newDeployment(chartName, releaseName, withDeploymentReplicas(2)),
		},
		{
			name: "sort volumes",
			assertOptions: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
				assert.WithSortObjectsByNameField(),
			},
			crateValues: []byte(volumes),
			want: newDeployment(chartName, releaseName,
				withDeploymentVolumes([]corev1.Volume{
					// reverse of dictionary order
					{
						Name: "bbb",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/bbb",
								Type: &hostPathDirectory,
							},
						},
					},
					{
						Name: "aaa",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/aaa",
								Type: &hostPathDirectory,
							},
						},
					},
				})),
		},
	}

	r := helmut.New()

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.crateValues != nil {
				path := createTempValuesFile(t, tt.crateValues)
				tt.options = append(tt.options, helmut.WithValues(path))
			}

			manifests, err := r.RenderTemplates(releaseName, filepath.Join("testdata", chartName), tt.options...)
			if err != nil {
				t.Fatalf("failed to render templates: %s", err)
			}

			assert.Contains(t, manifests, tt.want, tt.assertOptions...)
		})
	}
}

func newServiceAccount(
	chartName, releaseName string,
	options ...func(account *corev1.ServiceAccount),
) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", releaseName, chartName),
		},
	}

	for _, option := range options {
		option(sa)
	}

	return sa
}

func newService(chartName, releaseName string, options ...func(*corev1.Service)) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", releaseName, chartName),
			Labels: map[string]string{
				"app.kubernetes.io/instance": releaseName,
				"app.kubernetes.io/name":     chartName,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromString("http"),
					Protocol:   corev1.ProtocolTCP,
					Name:       "http",
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":     chartName,
				"app.kubernetes.io/instance": releaseName,
			},
		},
	}

	for _, option := range options {
		option(svc)
	}

	return svc
}

func withServiceLabels(labels map[string]string) func(*corev1.Service) {
	return func(svc *corev1.Service) {
		svc.Labels = labels
	}
}

func newDeployment(chartName, releaseName string, options ...func(*appsv1.Deployment)) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", releaseName, chartName),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":     chartName,
					"app.kubernetes.io/instance": releaseName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":     chartName,
						"app.kubernetes.io/instance": releaseName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: fmt.Sprintf("%s-%s", releaseName, chartName),
					Containers: []corev1.Container{
						{
							Name:            chartName,
							Image:           "nginx:1.16.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: int32(80),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.IntOrString{
											Type:   1,
											StrVal: "http",
										},
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.IntOrString{
											Type:   1,
											StrVal: "http",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, option := range options {
		option(deploy)
	}

	return deploy
}

func withDeploymentReplicas(replicas int32) func(*appsv1.Deployment) {
	return func(deploy *appsv1.Deployment) {
		deploy.Spec.Replicas = pointer.Int32Ptr(replicas)
	}
}

func withDeploymentVolumes(volumes []corev1.Volume) func(*appsv1.Deployment) {
	return func(deploy *appsv1.Deployment) {
		deploy.Spec.Template.Spec.Volumes = volumes
	}
}

func createTempValuesFile(t *testing.T, data []byte) string {
	t.Helper()

	dir := t.TempDir()

	f, err := os.CreateTemp(dir, "temp-values")
	if err != nil {
		t.Fatalf("failed to carete tempfile: %s", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write file: %s", err)
	}

	if err := f.Sync(); err != nil {
		t.Fatalf("failed to fsync: %s", err)
	}

	return f.Name()
}
