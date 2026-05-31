package play

import (
	"errors"
	"testing"
)

func TestHelpersCoverage_ResultMessage_Good(testingT *testing.T) {
	testingT.Parallel()

	if got := resultMessage(errors.New("boom")); got != "boom" {
		testingT.Fatalf("resultMessage(error) = %q, want boom", got)
	}
	if got := resultMessage("plain text"); got != "plain text" {
		testingT.Fatalf("resultMessage(string) = %q, want plain text", got)
	}
}

func TestHelpersCoverage_ResultMessage_Bad(testingT *testing.T) {
	testingT.Parallel()

	if got := resultMessage(nil); got != "process execution failed" {
		testingT.Fatalf("resultMessage(nil) = %q, want process execution failed", got)
	}
}

func TestHelpersCoverage_ResultMessage_Ugly(testingT *testing.T) {
	testingT.Parallel()

	// A non-error, non-string value falls through to a formatted representation.
	if got := resultMessage(42); got == "" {
		testingT.Fatal("resultMessage(int) returned empty string")
	}
}

func TestHelpersCoverage_ResultError_Good(testingT *testing.T) {
	testingT.Parallel()

	sentinel := errors.New("serialise failed")
	if got := resultError(sentinel); got != sentinel {
		testingT.Fatalf("resultError(error) = %v, want sentinel", got)
	}
}

func TestHelpersCoverage_ResultError_Ugly(testingT *testing.T) {
	testingT.Parallel()

	if got := resultError("not an error"); got != nil {
		testingT.Fatalf("resultError(string) = %v, want nil", got)
	}
	if got := resultError(nil); got != nil {
		testingT.Fatalf("resultError(nil) = %v, want nil", got)
	}
}

func TestHelpersCoverage_RetroArchCoreName_Good(testingT *testing.T) {
	testingT.Parallel()

	cases := map[string]string{
		"genesis":          "genesis_plus_gx",
		"snes":             "snes9x",
		"nes":              "nestopia",
		"game-boy":         "gambatte",
		"game-boy-color":   "gambatte",
		"game-boy-advance": "mgba",
		"gba":              "mgba",
	}
	for profile, want := range cases {
		got, ok := retroArchCoreName(profile)
		if !ok || got != want {
			testingT.Fatalf("retroArchCoreName(%q) = (%q, %t), want (%q, true)", profile, got, ok, want)
		}
	}
}

func TestHelpersCoverage_RetroArchCoreName_Ugly(testingT *testing.T) {
	testingT.Parallel()

	if got, ok := retroArchCoreName("unknown-profile"); ok || got != "" {
		testingT.Fatalf("retroArchCoreName(unknown) = (%q, %t), want (empty, false)", got, ok)
	}
}

func TestHelpersCoverage_ViceModel_Good(testingT *testing.T) {
	testingT.Parallel()

	cases := map[string]string{
		"commodore-64":  "c64pal",
		"c64":           "c64pal",
		"commodore-128": "c128pal",
		"c128":          "c128pal",
		"vic-20":        "vic20pal",
	}
	for profile, want := range cases {
		got, ok := viceModel(profile)
		if !ok || got != want {
			testingT.Fatalf("viceModel(%q) = (%q, %t), want (%q, true)", profile, got, ok, want)
		}
	}
}

func TestHelpersCoverage_ViceModel_Ugly(testingT *testing.T) {
	testingT.Parallel()

	if got, ok := viceModel("plus-4"); ok || got != "" {
		testingT.Fatalf("viceModel(unknown) = (%q, %t), want (empty, false)", got, ok)
	}
}

func TestHelpersCoverage_ProcessAccelerationOff_Good(testingT *testing.T) {
	testingT.Parallel()

	pipeline := FramePipeline{
		Primary:  acceleratedFrameProcessor{name: "metal", available: true},
		Fallback: acceleratedFrameProcessor{name: "cpu"},
	}
	result, err := pipeline.Process(validRGBAFrame(), FramePolicy{Mode: AccelerationOff})
	if err != nil {
		testingT.Fatalf("Process returned error: %v", err)
	}
	if result.Accelerated || !result.Fallback {
		testingT.Fatalf("Process with AccelerationOff expected fallback: %+v", result)
	}
	if result.Processor != "cpu" {
		testingT.Fatalf("unexpected processor: %q", result.Processor)
	}
}

func TestHelpersCoverage_ProcessAccelerationFailed_Bad(testingT *testing.T) {
	testingT.Parallel()

	pipeline := FramePipeline{
		Primary: acceleratedFrameProcessor{
			name:       "metal",
			available:  true,
			processErr: EngineError{Kind: "engine/test-failure", Message: "synthetic"},
		},
	}
	_, err := pipeline.Process(validRGBAFrame(), FramePolicy{Mode: AccelerationRequired})
	if err == nil {
		testingT.Fatal("Process expected an error when required acceleration fails")
	}
	pipelineError, ok := err.(PipelineError)
	if !ok {
		testingT.Fatalf("Process returned %T, want PipelineError", err)
	}
	if pipelineError.Kind != "frame/acceleration-failed" {
		testingT.Fatalf("unexpected pipeline error kind: %q", pipelineError.Kind)
	}
}

func TestHelpersCoverage_ProcessPrimaryErrorFallback_Ugly(testingT *testing.T) {
	testingT.Parallel()

	// Primary is available but errors; auto mode falls back to the CPU path.
	pipeline := FramePipeline{
		Primary: acceleratedFrameProcessor{
			name:       "metal",
			available:  true,
			processErr: EngineError{Kind: "engine/test-failure", Message: "synthetic"},
		},
		Fallback: acceleratedFrameProcessor{name: "cpu"},
	}
	result, err := pipeline.Process(validRGBAFrame(), FramePolicy{Mode: AccelerationAuto})
	if err != nil {
		testingT.Fatalf("Process returned error: %v", err)
	}
	if !result.Fallback || result.Processor != "cpu" {
		testingT.Fatalf("Process expected CPU fallback after primary error: %+v", result)
	}
}
