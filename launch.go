package play

// LaunchPlan describes how an engine intends to run a bundle.
type LaunchPlan struct {
	Engine           string
	Executable       string
	Arguments        []string
	WorkingDirectory string
	Entrypoint       string
	RuntimeConfig    string
	ReadPaths        []string
	WritePaths       []string
	NetworkAllowed   bool
}

// LaunchPlanner is implemented by engines that can turn a bundle into a launch plan.
type LaunchPlanner interface {
	PlanLaunch(bundle Bundle) (LaunchPlan, error)
}

// PlanLaunch asks an engine for a launch plan when it supports launch planning.
func PlanLaunch(bundle Bundle, engine Engine) (LaunchPlan, error) {
	if engine == nil {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/nil",
			Message: "engine is required",
		}
	}

	planner, ok := engine.(LaunchPlanner)
	if !ok {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/plan-unavailable",
			Name:    engine.Name(),
			Message: "engine does not provide a launch plan",
		}
	}

	return planner.PlanLaunch(bundle)
}

func clonePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	cloned := make([]string, len(paths))
	copy(cloned, paths)

	return cloned
}
