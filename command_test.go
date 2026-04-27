package play

import (
	"testing"

	"dappco.re/go/core"
)

func TestCommand_Commands_Good(testingT *testing.T) {
	testingT.Parallel()

	commands := Commands()
	if len(commands) != 6 {
		testingT.Fatalf("unexpected command count: %d", len(commands))
	}
	if commands[0] != CommandPlay {
		testingT.Fatalf("unexpected first command: %q", commands[0])
	}
}

func TestCommand_Commands_Bad(testingT *testing.T) {
	testingT.Parallel()

	commands := Commands()
	for _, command := range commands {
		if command == "" {
			testingT.Fatal("Commands returned an empty command name")
		}
	}
}

func TestCommand_Commands_Ugly(testingT *testing.T) {
	testingT.Parallel()

	first := Commands()
	second := Commands()

	first[0] = "changed"
	if second[0] != CommandPlay {
		testingT.Fatal("Commands should return an isolated slice on each call")
	}
}

func TestCommand_Register_Good(testingT *testing.T) {
	testingT.Parallel()

	c := core.New()
	Register(c)

	commands := c.Commands()
	if len(commands) != 6 {
		testingT.Fatalf("unexpected registered command count: %d", len(commands))
	}
	if commands[0] != CommandPlay {
		testingT.Fatalf("unexpected first registered command: %q", commands[0])
	}
}

func TestCommand_Register_Bad(testingT *testing.T) {
	testingT.Parallel()

	Register(nil)
}

func TestCommand_Register_Ugly(testingT *testing.T) {
	testingT.Parallel()

	c := core.New()
	Register(c)
	Register(c)

	if len(c.Commands()) != 6 {
		testingT.Fatalf("duplicate Register changed command count: %d", len(c.Commands()))
	}
}

func TestCommand_PlayRequest_Good(testingT *testing.T) {
	testingT.Parallel()

	result := cmdPlay(core.NewOptions(core.Option{Key: "name", Value: "mega-lo-mania"}))
	request := result.Value.(PlayRequest)

	if request.BundlePath != "mega-lo-mania" {
		testingT.Fatalf("unexpected play bundle path: %q", request.BundlePath)
	}
}

func TestCommand_ListRequest_Good(testingT *testing.T) {
	testingT.Parallel()

	result := cmdPlayList(core.NewOptions(
		core.Option{Key: "root", Value: "games"},
		core.Option{Key: "json", Value: true},
	))
	request := result.Value.(ListRequest)

	if request.Root != "games" {
		testingT.Fatalf("unexpected list root: %q", request.Root)
	}
	if !request.JSON {
		testingT.Fatal("expected list JSON option to be preserved")
	}
}

func TestCommand_VerifyRequest_Good(testingT *testing.T) {
	testingT.Parallel()

	result := cmdPlayVerify(core.NewOptions(core.Option{Key: "bundle", Value: "mega-lo-mania"}))
	request := result.Value.(VerifyRequest)

	if request.BundlePath != "mega-lo-mania" {
		testingT.Fatalf("unexpected verify bundle path: %q", request.BundlePath)
	}
}

func TestCommand_BundleRequest_Good(testingT *testing.T) {
	testingT.Parallel()

	artefactData := []byte("rom")
	engineData := []byte("engine")
	result := cmdPlayBundle(core.NewOptions(
		core.Option{Key: "name", Value: "mega-lo-mania"},
		core.Option{Key: "title", Value: "Mega lo Mania"},
		core.Option{Key: "author", Value: "Sensible Software"},
		core.Option{Key: "year", Value: 1991},
		core.Option{Key: "platform", Value: "sega-genesis"},
		core.Option{Key: "genre", Value: "strategy"},
		core.Option{Key: "licence", Value: "freeware"},
		core.Option{Key: "engine", Value: "retroarch"},
		core.Option{Key: "profile", Value: "genesis"},
		core.Option{Key: "acceleration", Value: "auto"},
		core.Option{Key: "filter", Value: "nearest"},
		core.Option{Key: "rom", Value: "rom/MegaLoMania.zip"},
		core.Option{Key: "artefact_data", Value: artefactData},
		core.Option{Key: "artefact_sha256", Value: hashHex(artefactData)},
		core.Option{Key: "artefact_size", Value: int64(len(artefactData))},
		core.Option{Key: "media_type", Value: "application/zip"},
		core.Option{Key: "source", Value: "rights-cleared"},
		core.Option{Key: "engine_binary", Value: "engine/retroarch"},
		core.Option{Key: "engine_binary_data", Value: engineData},
		core.Option{Key: "engine_binary_sha256", Value: hashHex(engineData)},
		core.Option{Key: "cpu_percent", Value: 75},
		core.Option{Key: "memory_bytes", Value: int64(268435456)},
		core.Option{Key: "distribution_mode", Value: "catalogue"},
		core.Option{Key: "byorom", Value: true},
		core.Option{Key: "entrypoint", Value: "rom/MegaLoMania.zip"},
		core.Option{Key: "config", Value: "emulator.yaml"},
		core.Option{Key: "chain", Value: "checksums.sha256"},
		core.Option{Key: "sbom", Value: "sbom.json"},
		core.Option{Key: "save_path", Value: "saves/"},
		core.Option{Key: "screenshot_path", Value: "screenshots/"},
	))
	request := result.Value.(BundleRequest)

	if request.Year != 1991 {
		testingT.Fatalf("unexpected bundle year: %d", request.Year)
	}
	if request.ArtefactPath != "rom/MegaLoMania.zip" {
		testingT.Fatalf("unexpected artefact path: %q", request.ArtefactPath)
	}
	if request.ResourceLimits.CPUPercent != 75 {
		testingT.Fatalf("unexpected CPU limit: %d", request.ResourceLimits.CPUPercent)
	}
	if request.ResourceLimits.MemoryBytes != 268435456 {
		testingT.Fatalf("unexpected memory limit: %d", request.ResourceLimits.MemoryBytes)
	}
	if !request.BYOROM {
		testingT.Fatal("expected BYOROM option to be preserved")
	}
	if string(request.ArtefactData) != "rom" {
		testingT.Fatalf("unexpected artefact data: %q", string(request.ArtefactData))
	}
	if string(request.EngineBinaryData) != "engine" {
		testingT.Fatalf("unexpected engine data: %q", string(request.EngineBinaryData))
	}
}

func TestCommand_BundleRequest_Ugly(testingT *testing.T) {
	testingT.Parallel()

	result := cmdPlayBundle(core.NewOptions(
		core.Option{Key: "artefact", Value: "rom/fallback.zip"},
		core.Option{Key: "cpu-percent", Value: 55},
		core.Option{Key: "memory-bytes", Value: int64(1024)},
	))
	request := result.Value.(BundleRequest)

	if request.ArtefactPath != "rom/fallback.zip" {
		testingT.Fatalf("unexpected fallback artefact path: %q", request.ArtefactPath)
	}
	if request.ResourceLimits.CPUPercent != 55 {
		testingT.Fatalf("unexpected fallback CPU limit: %d", request.ResourceLimits.CPUPercent)
	}
	if request.ResourceLimits.MemoryBytes != 1024 {
		testingT.Fatalf("unexpected fallback memory limit: %d", request.ResourceLimits.MemoryBytes)
	}
}
