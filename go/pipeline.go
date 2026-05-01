package play

// AccelerationMode describes how strongly GPU or accelerated processing is preferred.
type AccelerationMode string

const (
	AccelerationOff      AccelerationMode = "off"
	AccelerationAuto     AccelerationMode = "auto"
	AccelerationRequired AccelerationMode = "required"
)

func (mode AccelerationMode) valid() bool {
	switch mode {
	case "", AccelerationOff, AccelerationAuto, AccelerationRequired:
		return true
	default:
		return false
	}
}

// FramePolicy describes how a frame should be processed.
type FramePolicy struct {
	Mode         AccelerationMode
	Filter       FrameFilter
	TargetWidth  int
	TargetHeight int
}

// FrameResult describes a processed frame and the path used to produce it.
type FrameResult struct {
	Frame       FrameBuffer
	Processor   string
	Accelerated bool
	Fallback    bool
}

// FrameProcessor handles frame processing for scaling, conversion, or post-processing.
type FrameProcessor interface {
	Name() string
	Available() bool
	Supports(frame FrameBuffer, policy FramePolicy) bool
	Process(frame FrameBuffer, policy FramePolicy) (FrameBuffer, error)
}

// AccelerationDescriptor describes an engine's relationship with frame acceleration.
type AccelerationDescriptor struct {
	Mode             AccelerationMode
	PreferredFilters []FrameFilter
}

// AccelerationDescriber is implemented by engines that can report acceleration preferences.
type AccelerationDescriber interface {
	Acceleration() AccelerationDescriptor
}

// FramePipeline routes frames through an accelerated path with a safe fallback.
type FramePipeline struct {
	Primary  FrameProcessor
	Fallback FrameProcessor
}

// Process runs the frame through the configured pipeline using the supplied policy.
func (pipeline FramePipeline) Process(frame FrameBuffer, policy FramePolicy) (FrameResult, error) {
	issues := frame.Validate()
	if issues.HasIssues() {
		return FrameResult{}, issues
	}

	selectedMode := normaliseAccelerationMode(policy.Mode)
	fallback := pipeline.fallbackProcessor()
	if selectedMode == AccelerationOff {
		output, err := fallback.Process(frame, policy)
		if err != nil {
			return FrameResult{}, PipelineError{
				Kind:    "frame/fallback-failed",
				Message: err.Error(),
			}
		}

		return FrameResult{
			Frame:       output,
			Processor:   fallback.Name(),
			Accelerated: false,
			Fallback:    true,
		}, nil
	}

	if pipeline.Primary == nil || !pipeline.Primary.Available() || !pipeline.Primary.Supports(frame, policy) {
		if selectedMode == AccelerationRequired {
			return FrameResult{}, PipelineError{
				Kind:    "frame/acceleration-required",
				Message: "accelerated frame processing is required but unavailable",
			}
		}

		output, err := fallback.Process(frame, policy)
		if err != nil {
			return FrameResult{}, PipelineError{
				Kind:    "frame/fallback-failed",
				Message: err.Error(),
			}
		}

		return FrameResult{
			Frame:       output,
			Processor:   fallback.Name(),
			Accelerated: false,
			Fallback:    true,
		}, nil
	}

	output, err := pipeline.Primary.Process(frame, policy)
	if err == nil {
		return FrameResult{
			Frame:       output,
			Processor:   pipeline.Primary.Name(),
			Accelerated: true,
			Fallback:    false,
		}, nil
	}

	if selectedMode == AccelerationRequired {
		return FrameResult{}, PipelineError{
			Kind:    "frame/acceleration-failed",
			Message: err.Error(),
		}
	}

	fallbackOutput, fallbackErr := fallback.Process(frame, policy)
	if fallbackErr != nil {
		return FrameResult{}, PipelineError{
			Kind:    "frame/fallback-failed",
			Message: fallbackErr.Error(),
		}
	}

	return FrameResult{
		Frame:       fallbackOutput,
		Processor:   fallback.Name(),
		Accelerated: false,
		Fallback:    true,
	}, nil
}

func (pipeline FramePipeline) fallbackProcessor() FrameProcessor {
	if pipeline.Fallback != nil {
		return pipeline.Fallback
	}

	return identityFrameProcessor{}
}

func normaliseAccelerationMode(mode AccelerationMode) AccelerationMode {
	switch mode {
	case AccelerationOff, AccelerationRequired:
		return mode
	default:
		return AccelerationAuto
	}
}

// FramePolicy returns the runtime display intent declared by the manifest.
func (manifest Manifest) FramePolicy() FramePolicy {
	return FramePolicy{
		Mode:   normaliseAccelerationMode(manifest.Runtime.Acceleration),
		Filter: normaliseFrameFilter(manifest.Runtime.Filter),
	}
}

func normaliseFrameFilter(filter FrameFilter) FrameFilter {
	if !filter.valid() || filter == "" {
		return FrameFilterNone
	}

	return filter
}

type identityFrameProcessor struct{}

func (identityFrameProcessor) Name() string {
	return "identity"
}

func (identityFrameProcessor) Available() bool {
	return true
}

func (identityFrameProcessor) Supports(FrameBuffer, FramePolicy) bool {
	return true
}

func (identityFrameProcessor) Process(frame FrameBuffer, _ FramePolicy) (FrameBuffer, error) {
	return frame.Clone(), nil
}

// PipelineError reports frame-processing pipeline failures.
type PipelineError struct {
	Kind    string
	Message string
}

func (pipelineError PipelineError) Error() string {
	return pipelineError.Kind + ": " + pipelineError.Message
}
