package play

// PixelFormat describes a framebuffer pixel layout.
type PixelFormat string

const (
	PixelFormatRGBA8    PixelFormat = "rgba8"
	PixelFormatBGRA8    PixelFormat = "bgra8"
	PixelFormatRGB565   PixelFormat = "rgb565"
	PixelFormatXRGB8888 PixelFormat = "xrgb8888"
	PixelFormatIndexed8 PixelFormat = "indexed8"
)

// FrameFilter describes an optional frame-processing filter.
type FrameFilter string

const (
	FrameFilterNone     FrameFilter = "none"
	FrameFilterNearest  FrameFilter = "nearest"
	FrameFilterBilinear FrameFilter = "bilinear"
	FrameFilterScanline FrameFilter = "scanline"
	FrameFilterCRT      FrameFilter = "crt"
)

func (filter FrameFilter) valid() bool {
	switch filter {
	case "", FrameFilterNone, FrameFilterNearest, FrameFilterBilinear, FrameFilterScanline, FrameFilterCRT:
		return true
	default:
		return false
	}
}

// FrameBuffer describes an emulator or runtime video frame.
type FrameBuffer struct {
	Width  int
	Height int
	Stride int
	Format PixelFormat
	Data   []byte
}

// Clone returns a detached copy of the frame buffer.
func (frame FrameBuffer) Clone() FrameBuffer {
	return FrameBuffer{
		Width:  frame.Width,
		Height: frame.Height,
		Stride: frame.Stride,
		Format: frame.Format,
		Data:   cloneBytes(frame.Data),
	}
}

// Validate checks whether a frame buffer is structurally valid.
func (frame FrameBuffer) Validate() ValidationErrors {
	var issues ValidationErrors

	if frame.Width <= 0 {
		issues = append(issues, ValidationIssue{
			Code:    "frame/width-invalid",
			Field:   "width",
			Message: "frame width must be greater than zero",
		})
	}
	if frame.Height <= 0 {
		issues = append(issues, ValidationIssue{
			Code:    "frame/height-invalid",
			Field:   "height",
			Message: "frame height must be greater than zero",
		})
	}

	bytesPerPixel, validFormat := frame.Format.bytesPerPixel()
	if !validFormat {
		issues = append(issues, ValidationIssue{
			Code:    "frame/format-invalid",
			Field:   "format",
			Message: "frame pixel format is not supported",
		})
		return issues
	}

	minimumStride := frame.Width * bytesPerPixel
	if frame.Stride < minimumStride {
		issues = append(issues, ValidationIssue{
			Code:    "frame/stride-invalid",
			Field:   "stride",
			Message: "frame stride is smaller than the minimum row width",
		})
	}

	requiredBytes := frame.Stride * frame.Height
	if requiredBytes <= 0 {
		requiredBytes = minimumStride * frame.Height
	}
	if len(frame.Data) < requiredBytes {
		issues = append(issues, ValidationIssue{
			Code:    "frame/data-too-short",
			Field:   "data",
			Message: "frame data is shorter than width, height, and stride require",
		})
	}

	return issues
}

func (format PixelFormat) bytesPerPixel() (int, bool) {
	switch format {
	case PixelFormatRGBA8, PixelFormatBGRA8, PixelFormatXRGB8888:
		return 4, true
	case PixelFormatRGB565:
		return 2, true
	case PixelFormatIndexed8:
		return 1, true
	default:
		return 0, false
	}
}
