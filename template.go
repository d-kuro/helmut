// Package helmut provides functions to perform unit tests on the helm chart.
package helmut

import (
	"errors"
	"fmt"
	"sync"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// Renderer will perform the equivalent of the "helm template" command to render the manifests.
type Renderer struct {
	scheme *runtime.Scheme

	once sync.Once
}

// New creates and returns a new Renderer.
// option allows you to specify the scheme to be used by the Renderer.
// For example, specify the scheme in which the custom resource is registered.
func New(options ...SchemeOption) *Renderer {
	opts := &schemeOption{}

	for _, o := range options {
		o(opts)
	}

	if opts.Empty() {
		return &Renderer{scheme: defaultScheme}
	}

	return &Renderer{scheme: opts.scheme}
}

// init registers the default scheme to the Renderer if no scheme is specified.
func (r *Renderer) init() {
	if r.scheme == nil {
		r.scheme = defaultScheme
	}
}

// RenderTemplates will execute the equivalent of the "helm template" command and return the result.
func (r *Renderer) RenderTemplates(name, chart string, options ...Option) (*Manifests, error) {
	r.once.Do(r.init)

	opts := &option{}

	for _, o := range options {
		o(opts)
	}

	client := newClient(name, opts)
	valueOpts := &values.Options{}
	settings := cli.New()

	chartPath, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to find chart directory: %w", err)
	}

	providers := getter.All(settings)

	values, err := valueOpts.MergeValues(providers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %w", err)
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load the chart: %w", err)
	}

	release, err := client.Run(chartRequested, values)
	if err != nil {
		return nil, fmt.Errorf("failed to render templates: %w", err)
	}

	return r.SplitManifests(release.Manifest)
}

// SplitManifests parses one huge YAML and splits it into individual manifests.
func (r *Renderer) SplitManifests(data string) (*Manifests, error) {
	r.once.Do(r.init)

	manifests := NewManifests(WithScheme(r.scheme))

	codecFactory := serializer.NewCodecFactory(r.scheme)
	deserializer := codecFactory.UniversalDeserializer()

	for _, manifest := range releaseutil.SplitManifests(data) {
		object, gvk, err := deserializer.Decode([]byte(manifest), nil, nil)

		// If Scheme is not registered, try to convert to unstructured.
		if runtime.IsNotRegisteredError(err) {
			object, gvk, err = deserializer.Decode([]byte(manifest), nil, &unstructured.Unstructured{})
			if err != nil {
				return nil, fmt.Errorf("failed to decode manifest to *unstructured.Unstructured: %w", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}

		if gvk == nil {
			return nil, errors.New("could not get GetGroupVersionKind as a result of decoding manifest")
		}

		accessor, err := meta.Accessor(object)
		if err != nil {
			return nil, fmt.Errorf("object is not a `metav1.Object`: %w", err)
		}

		key := NewObjectKey(accessor.GetNamespace(), accessor.GetName(), *gvk)

		manifests.Store(key, object)
	}

	return manifests, nil
}

// newClient creates and returns a new helm client.
func newClient(name string, opts *option) *action.Install {
	client := action.NewInstall(&action.Configuration{})

	client.DryRun = true
	client.Namespace = opts.namespace
	client.ReleaseName = name
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.APIVersions = opts.apiVersions
	client.IncludeCRDs = opts.includeCRDs

	return client
}
