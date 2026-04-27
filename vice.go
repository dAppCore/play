package play

// VICEEngine adapts VICE for Commodore 8-bit bundles.
type VICEEngine struct {
	Binary       string
	BinarySHA256 string
}

// Name returns the engine identifier.
func (engine VICEEngine) Name() string {
	return "vice"
}

// Platforms returns platforms supported by the initial VICE scaffold.
func (engine VICEEngine) Platforms() []string {
	return []string{
		"commodore-64",
		"c64",
		"commodore-128",
		"c128",
		"vic-20",
	}
}

// Acceleration reports VICE's preferred frame-acceleration behaviour.
func (engine VICEEngine) Acceleration() AccelerationDescriptor {
	return AccelerationDescriptor{
		Mode: AccelerationAuto,
		PreferredFilters: []FrameFilter{
			FrameFilterNearest,
			FrameFilterCRT,
		},
	}
}

// Verify performs lightweight adapter validation.
func (engine VICEEngine) Verify() error {
	return requireEngineBinary(engine.Name(), engine.Binary)
}

// CodeIdentity returns the VICE runtime integrity identity.
func (engine VICEEngine) CodeIdentity() EngineCodeIdentity {
	return adapterCodeIdentity(engine.Name(), engine.Binary, engine.BinarySHA256)
}

// Run autostarts a Commodore artefact through VICE using Core's process primitive.
func (engine VICEEngine) Run(artefact string, config EngineConfig) error {
	if err := engine.Verify(); err != nil {
		return err
	}

	arguments, err := viceArguments(artefact, config.ConfigPath, config.Profile)
	if err != nil {
		return err
	}

	return runLaunchPlan(LaunchPlan{
		Engine:           engine.Name(),
		Executable:       engine.Binary,
		Arguments:        arguments,
		WorkingDirectory: ".",
		Entrypoint:       artefact,
		RuntimeConfig:    config.ConfigPath,
		NetworkAllowed:   config.NetworkAllowed,
	}, config)
}

// PlanLaunch builds a launch plan for a VICE-backed bundle.
func (engine VICEEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if err := validateAdapterBundle(engine.Name(), engine.Platforms(), bundle); err != nil {
		return LaunchPlan{}, err
	}

	entrypoint := manifestEntrypoint(bundle.Manifest)
	arguments, err := viceArguments(entrypoint, bundle.Manifest.Runtime.Config, bundle.Manifest.Runtime.Profile)
	if err != nil {
		return LaunchPlan{}, err
	}

	return adapterLaunchPlan(engine.Name(), engine.Binary, bundle, arguments), nil
}

func viceArguments(entrypoint string, configPath string, profile string) ([]string, error) {
	arguments := []string{}
	if configPath != "" {
		arguments = append(arguments, "-config", configPath)
	}
	if profile != "" {
		model, ok := viceModel(profile)
		if !ok {
			return nil, EngineError{
				Kind:    "engine/profile-unsupported",
				Name:    "vice",
				Message: "runtime profile is not supported by VICE",
			}
		}
		arguments = append(arguments, "-model", model)
	}
	arguments = append(arguments, "-autostart", entrypoint)

	return arguments, nil
}

func viceModel(profile string) (string, bool) {
	switch profile {
	case "commodore-64", "c64":
		return "c64pal", true
	case "commodore-128", "c128":
		return "c128pal", true
	case "vic-20":
		return "vic20pal", true
	default:
		return "", false
	}
}
