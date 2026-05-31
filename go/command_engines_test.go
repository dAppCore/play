package play

import (
	"testing"

	core "dappco.re/go"
)

func TestCommandEngines_CmdPlayEngines_Good(testingT *testing.T) {
	testingT.Parallel()

	if _, found := ResolveEngine("retroarch"); !found {
		if err := RegisterEngine(RetroArchEngine{Binary: "retroarch"}); err != nil {
			testingT.Fatalf("RegisterEngine returned error: %v", err)
		}
	}

	result := cmdPlayEngines(core.NewOptions())
	names, ok := result.Value.([]string)
	if !ok {
		testingT.Fatalf("cmdPlayEngines returned %T, want []string", result.Value)
	}

	found := false
	for _, name := range names {
		if name == "retroarch" {
			found = true
		}
	}
	if !found {
		testingT.Fatalf("cmdPlayEngines did not report retroarch: %v", names)
	}
}
