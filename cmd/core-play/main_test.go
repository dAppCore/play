package main

import "testing"

func TestMain_Parse_Good(testingT *testing.T) {
	testingT.Parallel()

	cases := []struct {
		args      []string
		operation operation
		bundle    string
		root      string
	}{
		{args: []string{"list", "--root", "tests/fixtures"}, operation: operationList, root: "tests/fixtures"},
		{args: []string{"verify", "sample-bundle"}, operation: operationVerify, bundle: "sample-bundle"},
		{args: []string{"info", "sample-bundle"}, operation: operationInfo, bundle: "sample-bundle"},
		{args: []string{"sample-bundle"}, operation: operationLaunch, bundle: "sample-bundle"},
	}

	for _, entry := range cases {
		parsed, err := parseInvocation(entry.args)
		if err != nil {
			testingT.Fatalf("parseInvocation returned error: %v", err)
		}
		if parsed.Operation != entry.operation {
			testingT.Fatalf("unexpected operation: %q", parsed.Operation)
		}
		if entry.bundle != "" && parsed.Bundle != entry.bundle {
			testingT.Fatalf("unexpected bundle: %q", parsed.Bundle)
		}
		if entry.root != "" && parsed.Root != entry.root {
			testingT.Fatalf("unexpected root: %q", parsed.Root)
		}
	}

	parsed, err := parseInvocation([]string{"bundle", "--name", "sample-bundle", "--rom", "rom.bin", "--cpu-percent", "75", "--memory-bytes", "268435456"})
	if err != nil {
		testingT.Fatalf("parseInvocation returned bundle error: %v", err)
	}
	if parsed.CPU != 75 || parsed.Memory != 268435456 {
		testingT.Fatalf("unexpected resource flags: cpu=%d memory=%d", parsed.CPU, parsed.Memory)
	}
}

func TestMain_Parse_Bad(testingT *testing.T) {
	testingT.Parallel()

	_, err := parseInvocation([]string{"list", "sample-bundle"})
	if err == nil {
		testingT.Fatal("parseInvocation expected an error for list bundle argument")
	}
}

func TestMain_Parse_Ugly(testingT *testing.T) {
	testingT.Parallel()

	_, err := parseInvocation([]string{"bundle", "--name", "sample-bundle"})
	if err == nil {
		testingT.Fatal("parseInvocation expected an error for missing rom flag")
	}
}
