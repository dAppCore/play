package play

import "testing"

func TestSandbox_PrepareSandbox_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	writer := newMemoryBundleWriter()
	policy, err := PrepareSandbox(bundle, "/home/core", writer)
	if err != nil {
		testingT.Fatalf("PrepareSandbox returned error: %v", err)
	}

	if policy.Root != "/home/core/.core/play/mega-lo-mania" {
		testingT.Fatalf("unexpected sandbox root: %q", policy.Root)
	}
	if !writer.hasDirectory("/home/core/.core/play/mega-lo-mania/saves") {
		testingT.Fatal("PrepareSandbox did not create save directory")
	}
}

func TestSandbox_PrepareSandbox_Bad(testingT *testing.T) {
	testingT.Parallel()

	_, err := PrepareSandbox(Bundle{}, "", newMemoryBundleWriter())
	if err == nil {
		testingT.Fatal("PrepareSandbox expected an error for missing home")
	}
}

func TestSandbox_PrepareSandbox_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	writer := newMemoryBundleWriter()
	writer.directoryErrors["/home/core/.core/play/mega-lo-mania/saves"] = WriteError{
		Kind:    "writer/denied",
		Path:    "saves",
		Message: "directory creation denied",
	}

	_, err = PrepareSandbox(bundle, "/home/core", writer)
	if err == nil {
		testingT.Fatal("PrepareSandbox expected an error when the writer fails")
	}
}

func TestSandbox_ValidateLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	policy := SandboxPolicy{
		ReadPaths: []string{
			"rom/",
		},
		WritePaths: []string{
			"saves/",
			"screenshots/",
		},
		NetworkAllowed: false,
	}
	plan := LaunchPlan{
		Entrypoint: "rom/rom.bin",
		ReadPaths: []string{
			"rom/",
		},
		WritePaths: []string{
			"saves/",
		},
		NetworkAllowed: false,
	}

	issues := policy.ValidateLaunch(plan)
	if issues.HasIssues() {
		testingT.Fatalf("ValidateLaunch returned issues: %v", issues)
	}
}

func TestSandbox_ValidateLaunch_Bad(testingT *testing.T) {
	testingT.Parallel()

	policy := SandboxPolicy{
		ReadPaths: []string{
			"rom/",
		},
		WritePaths: []string{
			"saves/",
		},
		NetworkAllowed: false,
	}
	plan := LaunchPlan{
		Entrypoint:     "rom/rom.bin",
		ReadPaths:      []string{"rom/"},
		WritePaths:     []string{"saves/"},
		NetworkAllowed: true,
	}

	issues := policy.ValidateLaunch(plan)
	if !hasIssueCode(issues, "sandbox/network-denied") {
		testingT.Fatalf("ValidateLaunch missing sandbox/network-denied issue: %v", issues)
	}
}

func TestSandbox_ValidateLaunch_Ugly(testingT *testing.T) {
	testingT.Parallel()

	policy := SandboxPolicy{
		ReadPaths: []string{
			"rom/",
		},
		WritePaths: []string{
			"saves/",
		},
		NetworkAllowed: false,
	}
	plan := LaunchPlan{
		Entrypoint: "payload/run.exe",
		ReadPaths: []string{
			"payload/",
		},
		WritePaths: []string{
			"cache/",
		},
		NetworkAllowed: false,
	}

	issues := policy.ValidateLaunch(plan)
	if !hasIssueCode(issues, "sandbox/read-denied") {
		testingT.Fatalf("ValidateLaunch missing sandbox/read-denied issue: %v", issues)
	}
	if !hasIssueCode(issues, "sandbox/write-denied") {
		testingT.Fatalf("ValidateLaunch missing sandbox/write-denied issue: %v", issues)
	}
}
