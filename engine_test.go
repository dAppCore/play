package play

import "testing"

func TestEngine_Register_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}})
	if err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	engine, found := registry.Resolve("retroarch")
	if !found {
		testingT.Fatal("Resolve did not return the registered engine")
	}
	if engine.Name() != "retroarch" {
		testingT.Fatalf("unexpected engine name: %q", engine.Name())
	}
}

func TestEngine_Register_Bad(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	err := registry.Register(nil)
	if err == nil {
		testingT.Fatal("Register expected an error for a nil engine")
	}

	engineError, ok := err.(EngineError)
	if !ok {
		testingT.Fatalf("Register returned %T, want EngineError", err)
	}
	if engineError.Kind != "engine/nil" {
		testingT.Fatalf("unexpected engine error kind: %q", engineError.Kind)
	}
}

func TestEngine_Register_Ugly(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	first := stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}
	second := stubEngine{name: "retroarch", platforms: []string{"dos"}}

	if err := registry.Register(first); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	err := registry.Register(second)
	if err == nil {
		testingT.Fatal("Register expected an error for duplicate names")
	}

	engineError, ok := err.(EngineError)
	if !ok {
		testingT.Fatalf("Register returned %T, want EngineError", err)
	}
	if engineError.Kind != "engine/duplicate" {
		testingT.Fatalf("unexpected engine error kind: %q", engineError.Kind)
	}
}

type stubEngine struct {
	name      string
	platforms []string
	verifyErr error
	codeHash  string
}

func (stub stubEngine) Name() string {
	return stub.name
}

func (stub stubEngine) Platforms() []string {
	return stub.platforms
}

func (stub stubEngine) Run(string, EngineConfig) error {
	return nil
}

func (stub stubEngine) Verify() error {
	return stub.verifyErr
}

func (stub stubEngine) CodeIdentity() EngineCodeIdentity {
	engineHash := stub.codeHash
	if engineHash == "" {
		engineHash = virtualEngineCodeSHA256(stub.name)
	}

	return EngineCodeIdentity{
		Name:   stub.name,
		Path:   stub.name,
		SHA256: engineHash,
	}
}
