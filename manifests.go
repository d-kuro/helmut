package helmut

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
)

// Manifests stores the rendered manifests.
type Manifests struct {
	objects map[ObjectKey]runtime.Object
	scheme  *runtime.Scheme

	mu   sync.RWMutex
	once sync.Once
}

// NewManifests creates and returns new Manifests.
// If SchemeOption is omitted, the default scheme will be used.
func NewManifests(options ...SchemeOption) *Manifests {
	m := &Manifests{}
	m.once.Do(m.init)

	opts := &schemeOption{}

	for _, o := range options {
		o(opts)
	}

	if opts.Empty() {
		return m
	}

	m.scheme = opts.scheme

	return m
}

// init is executed by sync.Once to initialize the struct.
func (m *Manifests) init() {
	if m.objects == nil {
		m.objects = make(map[ObjectKey]runtime.Object)
	}

	if m.scheme == nil {
		m.scheme = defaultScheme
	}
}

// Load returns the object stored in the manifests for a key, or nil if no value is present.
// The ok result indicates whether value was found in the manifests.
func (m *Manifests) Load(key ObjectKey) (runtime.Object, bool) {
	m.once.Do(m.init)

	m.mu.RLock()
	defer m.mu.RUnlock()

	obj, ok := m.objects[key]

	return obj, ok
}

// Delete deletes the object for a key.
func (m *Manifests) Delete(key ObjectKey) {
	m.once.Do(m.init)

	m.mu.Lock()
	delete(m.objects, key)
	m.mu.Unlock()
}

// Store sets the object for a key.
func (m *Manifests) Store(key ObjectKey, value runtime.Object) {
	m.once.Do(m.init)

	m.mu.Lock()
	m.objects[key] = value
	m.mu.Unlock()
}

// GetScheme returns the scheme.
func (m *Manifests) GetScheme() *runtime.Scheme {
	m.once.Do(m.init)

	return m.scheme
}

// Length returns the number of elements in the stored manifest.
func (m *Manifests) Length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.objects)
}

// GetKeys returns a list of keys for an object.
func (m *Manifests) GetKeys() []ObjectKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]ObjectKey, 0, m.Length())

	for key := range m.objects {
		keys = append(keys, key)
	}

	return keys
}
