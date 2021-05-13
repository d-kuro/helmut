package helmut

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// defaultScheme will be used by default if no scheme is specified in the option.
var defaultScheme = runtime.NewScheme()

// init registers a scheme to defaultScheme.
func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(defaultScheme))
	// Register the scheme for CRD.
	utilruntime.Must(apiextensionsv1.AddToScheme(defaultScheme))
	utilruntime.Must(apiextensionsv1beta1.AddToScheme(defaultScheme))
}

// schemeOption stores the scheme options.
type schemeOption struct {
	scheme *runtime.Scheme
}

// Empty returns true if the scheme is not registered.
func (o *schemeOption) Empty() bool {
	return o.scheme == nil
}

// SchemeOption is an option to specify the scheme.
// If not specified, the default scheme provided by client-go will be used.
type SchemeOption func(*schemeOption)

// WithScheme specifies the scheme.
func WithScheme(scheme *runtime.Scheme) SchemeOption {
	return func(o *schemeOption) {
		o.scheme = scheme
	}
}
