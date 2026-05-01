package main

import core "dappco.re/go"

func TestAX7_BundleWriter_EnsureDirectory_Good(testingT *core.T) {
	root := testingT.TempDir()
	writer := localBundleWriter{Root: root}
	err := writer.EnsureDirectory("bundle/rom")
	core.AssertNoError(testingT, err)
	core.AssertTrue(testingT, core.Stat(core.Path(root, "bundle/rom")).OK)
}

func TestAX7_BundleWriter_EnsureDirectory_Bad(testingT *core.T) {
	root := testingT.TempDir()
	writer := localBundleWriter{Root: root}
	err := writer.EnsureDirectory("bad\x00path")
	core.AssertError(testingT, err)
	core.AssertNotNil(testingT, err)
}

func TestAX7_BundleWriter_EnsureDirectory_Ugly(testingT *core.T) {
	root := testingT.TempDir()
	writer := localBundleWriter{Root: root}
	err := writer.EnsureDirectory("")
	core.AssertNoError(testingT, err)
	core.AssertTrue(testingT, core.Stat(root).OK)
}

func TestAX7_BundleWriter_WriteFile_Good(testingT *core.T) {
	root := testingT.TempDir()
	writer := localBundleWriter{Root: root}
	core.RequireNoError(testingT, writer.EnsureDirectory("bundle/rom"))
	err := writer.WriteFile("bundle/rom/game.bin", []byte("rom"))
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "rom", string(core.ReadFile(core.Path(root, "bundle/rom/game.bin")).Value.([]byte)))
}

func TestAX7_BundleWriter_WriteFile_Bad(testingT *core.T) {
	root := testingT.TempDir()
	writer := localBundleWriter{Root: root}
	err := writer.WriteFile("missing/game.bin", []byte("rom"))
	core.AssertError(testingT, err)
	core.AssertNotNil(testingT, err)
}

func TestAX7_BundleWriter_WriteFile_Ugly(testingT *core.T) {
	root := testingT.TempDir()
	writer := localBundleWriter{Root: root}
	err := writer.WriteFile("empty.bin", nil)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "", string(core.ReadFile(core.Path(root, "empty.bin")).Value.([]byte)))
}
