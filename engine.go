package play

import "sort"

// Engine defines a runnable STIM runtime adapter.
type Engine interface {
	Name() string
	Platforms() []string
	Verify() error
}

// Registry stores known engines by name.
type Registry struct {
	engines map[string]Engine
}

// NewRegistry creates an empty engine registry.
func NewRegistry() *Registry {
	return &Registry{
		engines: map[string]Engine{},
	}
}

// Register stores an engine by its declared name.
func (registry *Registry) Register(engine Engine) error {
	if registry == nil {
		return EngineError{
			Kind:    "engine/registry-missing",
			Message: "engine registry is required",
		}
	}
	if engine == nil {
		return EngineError{
			Kind:    "engine/nil",
			Message: "engine is required",
		}
	}

	name := engine.Name()
	if name == "" {
		return EngineError{
			Kind:    "engine/name-required",
			Message: "engine name is required",
		}
	}
	if _, exists := registry.engines[name]; exists {
		return EngineError{
			Kind:    "engine/duplicate",
			Name:    name,
			Message: "engine is already registered",
		}
	}

	registry.engines[name] = engine

	return nil
}

// Resolve returns a registered engine by name.
func (registry *Registry) Resolve(name string) (Engine, bool) {
	if registry == nil {
		return nil, false
	}
	engine, found := registry.engines[name]
	return engine, found
}

// Names returns registered engine names in a stable order.
func (registry *Registry) Names() []string {
	if registry == nil {
		return nil
	}
	names := make([]string, 0, len(registry.engines))
	for name := range registry.engines {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

var defaultRegistry = NewRegistry()

// RegisterEngine stores an engine in the package-level registry.
func RegisterEngine(engine Engine) error {
	return defaultRegistry.Register(engine)
}

// ResolveEngine returns an engine from the package-level registry.
func ResolveEngine(name string) (Engine, bool) {
	return defaultRegistry.Resolve(name)
}

// RegisteredEngines lists package-level engine names in a stable order.
func RegisteredEngines() []string {
	return defaultRegistry.Names()
}

// EngineError reports registry or engine verification failures.
type EngineError struct {
	Kind    string
	Name    string
	Message string
}

func (engineError EngineError) Error() string {
	if engineError.Name == "" {
		return engineError.Kind + ": " + engineError.Message
	}

	return engineError.Kind + ": " + engineError.Name + ": " + engineError.Message
}
