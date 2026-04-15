package play

// Core Play command names.
const (
	CommandPlay       = "play"
	CommandPlayList   = "play/list"
	CommandPlayVerify = "play/verify"
	CommandPlayBundle = "play/bundle"
)

// Commands returns known Core Play command names in a stable order.
func Commands() []string {
	return []string{
		CommandPlay,
		CommandPlayList,
		CommandPlayVerify,
		CommandPlayBundle,
	}
}

// PlayRequest describes a request to prepare a bundle for launch.
type PlayRequest struct {
	BundlePath string
}

// ListRequest describes a request to enumerate bundles.
type ListRequest struct {
	Root string
}

// VerifyRequest describes a request to verify a bundle.
type VerifyRequest struct {
	BundlePath string
}

// BundleRequest describes the input required to plan a new STIM bundle.
type BundleRequest struct {
	Name              string
	Title             string
	Author            string
	Year              int
	Platform          string
	Genre             string
	Licence           string
	Engine            string
	Profile           string
	Acceleration      AccelerationMode
	Filter            FrameFilter
	ArtefactPath      string
	ArtefactSHA256    string
	ArtefactSize      int64
	ArtefactMediaType string
	ArtefactSource    string
	DistributionMode  string
	BYOROM            bool
	Entrypoint        string
	RuntimeConfigPath string
	VerificationChain string
	SBOMPath          string
	SavePath          string
	ScreenshotPath    string
}
