package main

import (
	"io"
	"os"
	"path"

	"dappco.re/go/core"
	"dappco.re/go/play"
)

func runList(parsed invocation, out io.Writer) error {
	root := listBundleRoot(parsed)
	catalogue := play.Catalogue{
		Bundles:  os.DirFS(root),
		BasePath: root,
	}
	summaries, err := catalogue.Walk(".")
	if err != nil {
		return err
	}
	if parsed.JSON {
		return catalogue.PrintJSON(out, summaries)
	}

	return catalogue.Print(out, summaries)
}

func runVerify(parsed invocation, out io.Writer) error {
	root := bundleRoot(parsed)
	service := play.NewService(os.DirFS(root), nil)
	result, err := service.VerifyBundle(play.VerifyRequest{BundlePath: parsed.Bundle})
	if err != nil {
		return err
	}

	printShieldReport(out, result.Shield)
	if !result.Verified {
		return result.Issues
	}

	return nil
}

func runInfo(parsed invocation, out io.Writer) error {
	return printBundleFile(bundleRoot(parsed), parsed.Bundle, "manifest.yaml", out)
}

func runEngines(out io.Writer) error {
	core.Print(out, "ENGINE  PLATFORMS")
	for _, name := range play.RegisteredEngines() {
		engine, found := play.ResolveEngine(name)
		if !found {
			continue
		}
		core.Print(out, "%s  %s", name, core.Join(", ", engine.Platforms()...))
	}

	return nil
}

func printShieldReport(out io.Writer, report play.ShieldReport) {
	core.Print(out, "ShieldReport")
	core.Print(out, "SBOM: %s path=%s serial=%s", statusText(report.SBOM.Valid), report.SBOM.Path, report.SBOM.SerialNumber)
	core.Print(out, "Code: %s engine=%s expected=%s actual=%s", statusText(report.Code.OK), report.Code.Engine, report.Code.ExpectedSHA256, report.Code.ActualSHA256)
	core.Print(out, "Content: %s chain=%s artefact=%s", statusText(report.Content.OK), report.Content.ChainPath, report.Content.ArtefactPath)
	core.Print(out, "Threat: %s findings=%d", statusText(report.Threat.OK), len(report.Threat.Findings))
	core.Print(out, "OverallOK: %t", report.OverallOK)
	if !report.OverallOK {
		printIssues(out, report.Issues())
	}
}

func statusText(ok bool) string {
	if ok {
		return "[Y]"
	}

	return "[N]"
}

func listBundleRoot(parsed invocation) string {
	if parsed.Root != "" {
		return absoluteHostPath(parsed.Root)
	}
	if home := os.Getenv("CORE_PLAY_HOME"); home != "" {
		return path.Join(home, "bundles")
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "."
	}

	return path.Dir(workingDirectory)
}

func absoluteHostPath(root string) string {
	if path.IsAbs(root) {
		return path.Clean(root)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return root
	}

	return path.Clean(path.Join(workingDirectory, root))
}
