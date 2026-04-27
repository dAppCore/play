package play

import (
	"context"
	"io"
	"sort"

	"dappco.re/go/core"
)

// Engine defines a runnable STIM runtime adapter.
type Engine interface {
	Name() string
	Platforms() []string
	Run(artefact string, config EngineConfig) error
	Verify() error
}

// EngineConfig carries runtime context into an engine adapter.
type EngineConfig struct {
	Core             *core.Core
	Context          context.Context
	WorkingDirectory string
	ConfigPath       string
	Profile          string
	SaveRoot         string
	Resources        ResourceLimits
	NetworkAllowed   bool
	Output           io.Writer
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

func runLaunchPlan(plan LaunchPlan, config EngineConfig) error {
	if config.Core == nil {
		return EngineError{
			Kind:    "engine/process-unavailable",
			Name:    plan.Engine,
			Message: "core process primitive is required to run this engine",
		}
	}

	runContext := config.Context
	if runContext == nil {
		runContext = context.Background()
	}

	workingDirectory := plan.WorkingDirectory
	if config.WorkingDirectory != "" {
		workingDirectory = config.WorkingDirectory
	}

	result := config.Core.Process().RunIn(runContext, workingDirectory, plan.Executable, plan.Arguments...)
	if !result.OK {
		return EngineError{
			Kind:    "engine/run-failed",
			Name:    plan.Engine,
			Message: resultMessage(result.Value),
		}
	}

	if config.Output != nil {
		if output, ok := result.Value.(string); ok && output != "" {
			core.Print(config.Output, "%s", output)
		}
	}

	return nil
}

func resultMessage(value any) string {
	if value == nil {
		return "process execution failed"
	}
	if err, ok := value.(error); ok {
		return err.Error()
	}
	if message, ok := value.(string); ok {
		return message
	}

	return core.Sprint(value)
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
