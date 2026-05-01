package play

import "testing"

func TestFrame_Validate_Good(testingT *testing.T) {
	testingT.Parallel()

	frame := FrameBuffer{
		Width:  320,
		Height: 224,
		Stride: 1280,
		Format: PixelFormatRGBA8,
		Data:   make([]byte, 1280*224),
	}

	issues := frame.Validate()
	if issues.HasIssues() {
		testingT.Fatalf("Validate returned issues: %v", issues)
	}
}

func TestFrame_Validate_Bad(testingT *testing.T) {
	testingT.Parallel()

	frame := FrameBuffer{
		Width:  0,
		Height: 0,
		Stride: 0,
		Format: PixelFormat("unknown"),
	}

	issues := frame.Validate()
	if !issues.HasIssues() {
		testingT.Fatal("Validate expected issues for an invalid frame")
	}
	if !hasIssueCode(issues, "frame/width-invalid") {
		testingT.Fatal("Validate missing frame/width-invalid issue")
	}
	if !hasIssueCode(issues, "frame/format-invalid") {
		testingT.Fatal("Validate missing frame/format-invalid issue")
	}
}

func TestFrame_Validate_Ugly(testingT *testing.T) {
	testingT.Parallel()

	frame := FrameBuffer{
		Width:  320,
		Height: 224,
		Stride: 100,
		Format: PixelFormatRGB565,
		Data:   make([]byte, 10),
	}

	issues := frame.Validate()
	if !hasIssueCode(issues, "frame/stride-invalid") {
		testingT.Fatal("Validate missing frame/stride-invalid issue")
	}
	if !hasIssueCode(issues, "frame/data-too-short") {
		testingT.Fatal("Validate missing frame/data-too-short issue")
	}
}
