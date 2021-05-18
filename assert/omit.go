package assert

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// omitMetadata will omit metadata based on options to ignore diffs.
// Since the object will be edited, it is recommended to use the deep copied object.
func omitMetadata(object runtime.Object, option *ignoreOption) runtime.Object {
	accessor, err := meta.Accessor(object)
	if err != nil {
		return object
	}

	if option.allHelmManagedLabels {
		omitAllHelmManagedLabels(accessor)
	}

	if len(option.labels) != 0 {
		omitLabels(accessor, option.labels)
	}

	if len(option.annotations) != 0 {
		omitAnnotations(accessor, option.annotations)
	}

	return object
}

// omitLabels will omit the specified labels.
func omitLabels(object metav1.Object, keys []string) {
	labels := object.GetLabels()

	for _, key := range keys {
		delete(labels, key)
	}

	if len(labels) == 0 {
		labels = nil
	}

	object.SetLabels(labels)
}

// omitAnnotations will omit the specified annotation.
func omitAnnotations(object metav1.Object, keys []string) {
	annotations := object.GetAnnotations()

	for _, key := range keys {
		delete(annotations, key)
	}

	if len(annotations) == 0 {
		annotations = nil
	}

	object.SetAnnotations(annotations)
}

// omitAllHelmManagedLabels will omit the labels used by helm.
func omitAllHelmManagedLabels(object metav1.Object) {
	labels := object.GetLabels()

	helmLabels := [...]helmManagedLabel{
		labelAppName,
		labelAppManagedBy,
		labelAppInstance,
		labelAppVersion,
		labelAppComponent,
		labelAppPartOf,
		labelHelmChart,
	}

	for _, helmLabel := range helmLabels {
		delete(labels, helmLabel.String())
	}

	if len(labels) == 0 {
		labels = nil
	}

	object.SetLabels(labels)
}
