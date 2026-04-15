package play

import "testing"

func TestPipeline_Process_Good(testingT *testing.T) {
	testingT.Parallel()

	pipeline := FramePipeline{
		Primary: acceleratedFrameProcessor{
			name:      "metal",
			available: true,
		},
	}

	result, err := pipeline.Process(validRGBAFrame(), FramePolicy{
		Mode:   AccelerationAuto,
		Filter: FrameFilterNearest,
	})
	if err != nil {
		testingT.Fatalf("Process returned error: %v", err)
	}

	if !result.Accelerated {
		testingT.Fatalf("Process expected accelerated result: %+v", result)
	}
	if result.Processor != "metal" {
		testingT.Fatalf("unexpected processor: %q", result.Processor)
	}
}

func TestPipeline_Process_Bad(testingT *testing.T) {
	testingT.Parallel()

	pipeline := FramePipeline{
		Primary: acceleratedFrameProcessor{
			name:      "metal",
			available: false,
		},
	}

	_, err := pipeline.Process(validRGBAFrame(), FramePolicy{
		Mode: AccelerationRequired,
	})
	if err == nil {
		testingT.Fatal("Process expected an error when acceleration is required")
	}

	pipelineError, ok := err.(PipelineError)
	if !ok {
		testingT.Fatalf("Process returned %T, want PipelineError", err)
	}
	if pipelineError.Kind != "frame/acceleration-required" {
		testingT.Fatalf("unexpected pipeline error kind: %q", pipelineError.Kind)
	}
}

func TestPipeline_Process_Ugly(testingT *testing.T) {
	testingT.Parallel()

	pipeline := FramePipeline{
		Primary: acceleratedFrameProcessor{
			name:      "metal",
			available: false,
			processErr: EngineError{
				Kind:    "engine/test-failure",
				Message: "synthetic failure",
			},
		},
		Fallback: acceleratedFrameProcessor{
			name: "cpu",
		},
	}

	result, err := pipeline.Process(validRGBAFrame(), FramePolicy{
		Mode: AccelerationAuto,
	})
	if err != nil {
		testingT.Fatalf("Process returned error: %v", err)
	}

	if result.Accelerated {
		testingT.Fatalf("Process expected fallback result: %+v", result)
	}
	if !result.Fallback {
		testingT.Fatalf("Process expected fallback flag: %+v", result)
	}
	if result.Processor != "cpu" {
		testingT.Fatalf("unexpected fallback processor: %q", result.Processor)
	}
}

func TestPipeline_FramePolicyFromManifest_Good(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	policy := manifest.FramePolicy()
	if policy.Mode != AccelerationAuto {
		testingT.Fatalf("unexpected policy mode: %q", policy.Mode)
	}
	if policy.Filter != FrameFilterNearest {
		testingT.Fatalf("unexpected policy filter: %q", policy.Filter)
	}
}

func validRGBAFrame() FrameBuffer {
	return FrameBuffer{
		Width:  320,
		Height: 224,
		Stride: 1280,
		Format: PixelFormatRGBA8,
		Data:   make([]byte, 1280*224),
	}
}

type acceleratedFrameProcessor struct {
	name       string
	available  bool
	processErr error
}

func (processor acceleratedFrameProcessor) Name() string {
	return processor.name
}

func (processor acceleratedFrameProcessor) Available() bool {
	return processor.available
}

func (processor acceleratedFrameProcessor) Supports(FrameBuffer, FramePolicy) bool {
	return true
}

func (processor acceleratedFrameProcessor) Process(frame FrameBuffer, _ FramePolicy) (FrameBuffer, error) {
	if processor.processErr != nil {
		return FrameBuffer{}, processor.processErr
	}

	return frame.Clone(), nil
}
