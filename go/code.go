package play

// CodeResult reports whether runtime engine code still matches the manifest pin.
type CodeResult struct {
	Engine         string
	Path           string
	ExpectedSHA256 string
	ActualSHA256   string
	Recorded       bool
	Match          bool
	OK             bool
	Issues         ValidationErrors
}

// EngineCodeIdentity describes the currently selected runtime engine code.
type EngineCodeIdentity struct {
	Name   string
	Path   string
	SHA256 string
}

// EngineCodeProvider is implemented by engines that expose an integrity identity.
type EngineCodeProvider interface {
	CodeIdentity() EngineCodeIdentity
}

func verifyCode(bundle Bundle, registry *Registry) CodeResult {
	result := CodeResult{
		Engine: bundle.Manifest.Runtime.Engine,
		Path:   bundle.Manifest.Verification.Engine.Path,
	}
	if registry == nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "engine/registry-missing",
			Field:   "runtime.engine",
			Message: "engine registry is required",
		})
		result.OK = false
		return result
	}

	engine, found := registry.Resolve(bundle.Manifest.Runtime.Engine)
	if !found {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "engine/unavailable",
			Field:   "runtime.engine",
			Message: "runtime engine is not registered",
		})
		result.OK = false
		return result
	}
	if err := engine.Verify(); err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "engine/verify-failed",
			Field:   "runtime.engine",
			Message: err.Error(),
		})
		result.OK = false
		return result
	}

	pin := bundle.Manifest.Verification.Engine
	result.ExpectedSHA256 = pin.SHA256
	result.Recorded = pin.SHA256 != ""
	if !result.Recorded {
		result.OK = true
		return result
	}

	actual := hashPinnedEngine(bundle, engine, pin)
	result.ActualSHA256 = actual.SHA256
	if result.Path == "" {
		result.Path = actual.Path
	}

	if actual.SHA256 == "" {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "code/hash-unavailable",
			Field:   "verification.engine.sha256",
			Message: "engine hash could not be recalculated",
		})
		result.OK = false
		return result
	}
	if actual.SHA256 != pin.SHA256 {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "code/hash-mismatch",
			Field:   "verification.engine.sha256",
			Message: "engine hash does not match the manifest",
		})
		result.OK = false
		return result
	}

	result.Match = true
	result.OK = true
	return result
}

func hashPinnedEngine(bundle Bundle, engine Engine, pin CodePin) EngineCodeIdentity {
	if pin.Path != "" && validBundlePath(pin.Path) {
		if data, ok := bundleFileData(bundle, pin.Path); ok {
			return EngineCodeIdentity{
				Name:   defaultString(pin.Name, engine.Name()),
				Path:   pin.Path,
				SHA256: sha256Hex(data),
			}
		}
	}

	if provider, ok := engine.(EngineCodeProvider); ok {
		return provider.CodeIdentity()
	}

	return EngineCodeIdentity{}
}

func plannedCodePin(request BundleRequest, engineName string) CodePin {
	if request.EngineBinarySHA256 != "" || len(request.EngineBinaryData) > 0 {
		enginePath := defaultString(request.EngineBinaryPath, engineName)
		engineHash := request.EngineBinarySHA256
		if len(request.EngineBinaryData) > 0 {
			engineHash = sha256Hex(request.EngineBinaryData)
		}

		return CodePin{
			Name:   engineName,
			Path:   enginePath,
			SHA256: engineHash,
		}
	}

	return CodePin{
		Name:   engineName,
		Path:   engineName,
		SHA256: virtualEngineCodeSHA256(engineName),
	}
}

func virtualEngineCodeSHA256(engineName string) string {
	return sha256Hex([]byte("core-play-engine:" + engineName))
}
