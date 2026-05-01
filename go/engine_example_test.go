package play

func ExampleNewRegistry() {
	_ = NewRegistry
}

func ExampleRegistry_Register() {
	_ = (*Registry).Register
}

func ExampleRegistry_Resolve() {
	_ = (*Registry).Resolve
}

func ExampleRegistry_Names() {
	_ = (*Registry).Names
}

func ExampleRegisterEngine() {
	_ = RegisterEngine
}

func ExampleResolveEngine() {
	_ = ResolveEngine
}

func ExampleRegisteredEngines() {
	_ = RegisteredEngines
}

func ExampleEngineError_Error() {
	_ = (*EngineError).Error
}
