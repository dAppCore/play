package main

import (
	"path"

	core "dappco.re/go"
	"dappco.re/go/play"
)

func TestParse_Invocation_Good(testingT *core.T) {
	parsed, err := parseInvocation([]string{"verify", "--root", "/bundles", "game"})
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, operationVerify, parsed.Operation)
	core.AssertEqual(testingT, "/bundles", parsed.Root)
	core.AssertEqual(testingT, "game", parsed.Bundle)
}

func TestParse_Invocation_Bad(testingT *core.T) {
	_, err := parseInvocation([]string{"list", "unexpected", "extra"})
	core.AssertError(testingT, err)
}

func TestParse_Invocation_Ugly(testingT *core.T) {
	// No arguments defaults to a launch of the current directory.
	parsed, err := parseInvocation(nil)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, operationLaunch, parsed.Operation)
	core.AssertEqual(testingT, ".", parsed.Bundle)
}

func TestParse_PlayAlias_Good(testingT *core.T) {
	for _, args := range [][]string{
		{"play", "list"},
		{"play", "--list"},
	} {
		parsed, err := parseInvocation(args)
		core.AssertNoError(testingT, err)
		core.AssertEqual(testingT, operationList, parsed.Operation)
	}
}

func TestParse_PlayAlias_Bad(testingT *core.T) {
	_, err := parseInvocation([]string{"play", "list", "unexpected"})
	core.AssertError(testingT, err)
}

func TestParse_PlayAlias_Ugly(testingT *core.T) {
	// A bare "play" alias with no sub-command launches the current directory.
	parsed, err := parseInvocation([]string{"play"})
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, operationLaunch, parsed.Operation)
	core.AssertEqual(testingT, ".", parsed.Bundle)
}

func TestParse_PlayAliasVariants_Good(testingT *core.T) {
	cases := map[string]operation{
		"verify":        operationVerify,
		"--verify":      operationVerify,
		"shield-verify": operationShieldVerify,
		"info":          operationInfo,
		"--info":        operationInfo,
		"engines":       operationEngines,
	}
	for sub, want := range cases {
		parsed, err := parseInvocation([]string{"play", sub})
		core.AssertNoError(testingT, err)
		core.AssertEqual(testingT, want, parsed.Operation)
	}
}

func TestParse_PlayAliasBundle_Good(testingT *core.T) {
	parsed, err := parseInvocation([]string{"play", "bundle", "--name", "game", "--rom", "rom/game.zip"})
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, operationBundle, parsed.Operation)
	core.AssertEqual(testingT, "game", parsed.Name)
	// Title defaults to the bundle name when not supplied.
	core.AssertEqual(testingT, "game", parsed.Title)
}

func TestParse_PlayAliasLaunch_Good(testingT *core.T) {
	parsed, err := parseInvocation([]string{"play", "some-game"})
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, operationLaunch, parsed.Operation)
	core.AssertEqual(testingT, "some-game", parsed.Bundle)
}

func TestParse_Bundle_Bad(testingT *core.T) {
	// Missing required name surfaces a parse error.
	_, err := parseInvocation([]string{"bundle", "--rom", "rom/game.zip"})
	core.AssertError(testingT, err)
}

func TestParse_BundleMissingROM_Bad(testingT *core.T) {
	_, err := parseInvocation([]string{"bundle", "--name", "game"})
	core.AssertError(testingT, err)
}

func TestParse_BundlePositional_Bad(testingT *core.T) {
	_, err := parseInvocation([]string{"bundle", "--name", "game", "--rom", "r.zip", "stray"})
	core.AssertError(testingT, err)
}

func TestHelper_BundleRoot_Good(testingT *core.T) {
	core.AssertEqual(testingT, "/explicit", bundleRoot(invocation{Root: "/explicit"}))
}

func TestHelper_BundleRoot_Bad(testingT *core.T) {
	// With no flag and no environment override the working directory is used.
	testingT.Setenv("CORE_PLAY_ROOT", "")
	core.AssertEqual(testingT, ".", bundleRoot(invocation{}))
}

func TestHelper_BundleRoot_Ugly(testingT *core.T) {
	testingT.Setenv("CORE_PLAY_ROOT", "/from-env")
	core.AssertEqual(testingT, "/from-env", bundleRoot(invocation{}))
}

func TestHelper_PlayHome_Good(testingT *core.T) {
	testingT.Setenv("CORE_PLAY_HOME", "/home-env")
	home, err := playHome()
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "/home-env", home)
}

func TestHelper_PlayHome_Ugly(testingT *core.T) {
	testingT.Setenv("CORE_PLAY_HOME", "")
	home, err := playHome()
	core.AssertNoError(testingT, err)
	core.AssertNotEmpty(testingT, home)
}

func TestHelper_ListBundleRoot_Good(testingT *core.T) {
	testingT.Setenv("CORE_PLAY_HOME", "/home-env")
	core.AssertEqual(testingT, "/home-env/bundles", listBundleRoot(invocation{}))
}

func TestHelper_ListBundleRoot_Bad(testingT *core.T) {
	core.AssertEqual(testingT, "/abs/root", listBundleRoot(invocation{Root: "/abs/root"}))
}

func TestHelper_ListBundleRoot_Ugly(testingT *core.T) {
	testingT.Setenv("CORE_PLAY_HOME", "")
	// With neither flag nor home set, the parent of the working directory is used.
	root := listBundleRoot(invocation{})
	core.AssertNotEmpty(testingT, root)
}

func TestHelper_AbsoluteHostPath_Good(testingT *core.T) {
	core.AssertEqual(testingT, "/clean/path", absoluteHostPath("/clean/../clean/path"))
}

func TestHelper_AbsoluteHostPath_Ugly(testingT *core.T) {
	// A relative path is joined onto the working directory and made absolute.
	resolved := absoluteHostPath("relative/sub")
	core.AssertTrue(testingT, path.IsAbs(resolved))
}

func TestHelper_OutputPath_Good(testingT *core.T) {
	core.AssertEqual(testingT, "root/bundle/file", outputPath("root", "bundle/file"))
}

func TestHelper_OutputPath_Bad(testingT *core.T) {
	core.AssertEqual(testingT, "bundle/file", outputPath("", "bundle/file"))
}

func TestHelper_OutputPath_Ugly(testingT *core.T) {
	core.AssertEqual(testingT, "bundle/file", outputPath(".", "bundle/file"))
}

func TestHelper_DefaultText_Good(testingT *core.T) {
	core.AssertEqual(testingT, "value", defaultText("value", "fallback"))
}

func TestHelper_DefaultText_Ugly(testingT *core.T) {
	core.AssertEqual(testingT, "fallback", defaultText("", "fallback"))
}

func TestHelper_PrintIssues_Good(testingT *core.T) {
	out := core.NewBuffer()
	printIssues(out, play.ValidationErrors{
		play.ValidationIssue{Code: "engine/unavailable", Message: "no engine"},
	})
	core.AssertContains(testingT, out.String(), "engine/unavailable")
}

func TestHelper_HashBytes_Good(testingT *core.T) {
	// SHA-256 of an empty input is a known constant.
	core.AssertEqual(testingT,
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		hashBytes(nil))
}
