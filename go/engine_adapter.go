package play

func requireEngineBinary(engineName string, binary string) error {
	if binary == "" {
		return EngineError{
			Kind:    "engine/binary-required",
			Name:    engineName,
			Message: "binary path is required",
		}
	}

	return nil
}

func adapterCodeIdentity(engineName string, binary string, binarySHA256 string) EngineCodeIdentity {
	enginePath := defaultString(binary, engineName)
	engineHash := binarySHA256
	if engineHash == "" {
		engineHash = virtualEngineCodeSHA256(engineName)
	}

	return EngineCodeIdentity{
		Name:   engineName,
		Path:   enginePath,
		SHA256: engineHash,
	}
}

func validateAdapterBundle(engineName string, platforms []string, bundle Bundle) error {
	if bundle.Manifest.Runtime.Engine != engineName {
		return EngineError{
			Kind:    "engine/runtime-mismatch",
			Name:    engineName,
			Message: "bundle runtime does not match the selected engine",
		}
	}
	if !supportsPlatform(platforms, bundle.Manifest.Platform) {
		return EngineError{
			Kind:    "engine/platform-unsupported",
			Name:    engineName,
			Message: "bundle platform is not supported by the selected engine",
		}
	}

	return nil
}

func manifestEntrypoint(manifest Manifest) string {
	if manifest.Runtime.Entrypoint != "" {
		return manifest.Runtime.Entrypoint
	}

	return manifest.Artefact.Path
}

func adapterLaunchPlan(engineName string, binary string, bundle Bundle, arguments []string) LaunchPlan {
	entrypoint := manifestEntrypoint(bundle.Manifest)

	return LaunchPlan{
		Engine:           engineName,
		Executable:       binary,
		Arguments:        clonePaths(arguments),
		WorkingDirectory: ".",
		Entrypoint:       entrypoint,
		RuntimeConfig:    bundle.Manifest.Runtime.Config,
		ReadPaths:        manifestLaunchReadPaths(bundle.Manifest),
		WritePaths:       manifestLaunchWritePaths(bundle.Manifest),
		Resources:        bundle.Manifest.Resources,
		NetworkAllowed:   bundle.Manifest.Permissions.Network,
	}
}
