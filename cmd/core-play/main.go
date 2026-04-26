package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"os"
	"path"

	"dappco.re/go/core"
	"dappco.re/go/play"
)

type operation string

const (
	operationLaunch operation = "launch"
	operationList   operation = "list"
	operationVerify operation = "verify"
	operationInfo   operation = "info"
	operationBundle operation = "bundle"
)

type invocation struct {
	Operation operation
	Root      string
	Bundle    string
	Name      string
	Title     string
	Author    string
	Year      int
	Platform  string
	Genre     string
	Licence   string
	Engine    string
	Profile   string
	ROM       string
	Source    string
	BYOROM    bool
}

func main() {
	c := core.New(core.WithOption("name", "core-play"))
	play.Register(c)

	if err := run(context.Background(), c, os.Args[1:], os.Stdout); err != nil {
		core.Error("play.main", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, c *core.Core, args []string, out io.Writer) error {
	parsed, err := parseInvocation(args)
	if err != nil {
		return err
	}

	switch parsed.Operation {
	case operationList:
		return runList(parsed, out)
	case operationVerify:
		return runVerify(parsed, out)
	case operationInfo:
		return runInfo(parsed, out)
	case operationBundle:
		return runBundle(parsed, out)
	default:
		return runLaunch(ctx, c, parsed, out)
	}
}

func parseInvocation(args []string) (invocation, error) {
	if len(args) == 0 {
		return invocation{Operation: operationLaunch, Bundle: "."}, nil
	}

	switch args[0] {
	case "play":
		return parsePlayAlias(args[1:])
	case "list", "play/list":
		return parseList(args[1:])
	case "verify", "play/verify":
		return parseNamed(operationVerify, args[1:])
	case "info":
		return parseNamed(operationInfo, args[1:])
	case "bundle", "play/bundle":
		return parseBundle(args[1:])
	default:
		return parseLaunch(args)
	}
}

func parsePlayAlias(args []string) (invocation, error) {
	if len(args) == 0 {
		return invocation{Operation: operationLaunch, Bundle: "."}, nil
	}
	switch args[0] {
	case "--list":
		return parseList(args[1:])
	case "--verify":
		return parseNamed(operationVerify, args[1:])
	case "--info":
		return parseNamed(operationInfo, args[1:])
	default:
		return parseLaunch(args)
	}
}

func parseList(args []string) (invocation, error) {
	flags := newFlagSet("list")
	root := flags.String("root", "", "bundle root")
	if err := flags.Parse(args); err != nil {
		return invocation{}, err
	}
	if flags.NArg() > 0 {
		return invocation{}, core.E("play.parse", "list does not accept a bundle argument", nil)
	}

	return invocation{Operation: operationList, Root: *root}, nil
}

func parseNamed(op operation, args []string) (invocation, error) {
	flags := newFlagSet(string(op))
	root := flags.String("root", "", "bundle root")
	if err := flags.Parse(args); err != nil {
		return invocation{}, err
	}
	if flags.NArg() > 1 {
		return invocation{}, core.E("play.parse", "expected at most one bundle argument", nil)
	}

	bundle := "."
	if flags.NArg() == 1 {
		bundle = flags.Arg(0)
	}

	return invocation{Operation: op, Root: *root, Bundle: bundle}, nil
}

func parseLaunch(args []string) (invocation, error) {
	flags := newFlagSet("launch")
	root := flags.String("root", "", "bundle root")
	if err := flags.Parse(args); err != nil {
		return invocation{}, err
	}
	if flags.NArg() > 1 {
		return invocation{}, core.E("play.parse", "expected at most one bundle argument", nil)
	}

	bundle := "."
	if flags.NArg() == 1 {
		bundle = flags.Arg(0)
	}

	return invocation{Operation: operationLaunch, Root: *root, Bundle: bundle}, nil
}

func parseBundle(args []string) (invocation, error) {
	flags := newFlagSet("bundle")
	root := flags.String("root", "", "output root")
	name := flags.String("name", "", "bundle name")
	title := flags.String("title", "", "bundle title")
	author := flags.String("author", "", "bundle author")
	year := flags.Int("year", 0, "release year")
	platform := flags.String("platform", "", "platform")
	genre := flags.String("genre", "", "genre")
	licence := flags.String("licence", "freeware", "licence")
	engine := flags.String("engine", "", "engine")
	profile := flags.String("profile", "", "runtime profile")
	rom := flags.String("rom", "", "artefact path")
	source := flags.String("source", "", "artefact source")
	byorom := flags.Bool("byorom", false, "BYOROM mode")
	if err := flags.Parse(args); err != nil {
		return invocation{}, err
	}
	if flags.NArg() > 0 {
		return invocation{}, core.E("play.parse", "bundle accepts flags only", nil)
	}
	if *name == "" {
		return invocation{}, core.E("play.parse", "bundle name is required", nil)
	}
	if *rom == "" {
		return invocation{}, core.E("play.parse", "rom path is required", nil)
	}

	return invocation{
		Operation: operationBundle,
		Root:      *root,
		Name:      *name,
		Title:     defaultText(*title, *name),
		Author:    *author,
		Year:      *year,
		Platform:  *platform,
		Genre:     *genre,
		Licence:   *licence,
		Engine:    *engine,
		Profile:   *profile,
		ROM:       *rom,
		Source:    *source,
		BYOROM:    *byorom,
	}, nil
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

func runList(parsed invocation, out io.Writer) error {
	root := bundleRoot(parsed)
	service := play.NewService(os.DirFS(root), nil)
	summaries, err := service.ListBundles(play.ListRequest{Root: "."})
	if err != nil {
		return err
	}

	for _, summary := range summaries {
		core.Print(out, "%s\t%s\t%s\t%s", summary.Name, summary.Title, summary.Platform, summary.Engine)
	}

	return nil
}

func runVerify(parsed invocation, out io.Writer) error {
	root := bundleRoot(parsed)
	bundle, err := play.LoadBundle(os.DirFS(root), parsed.Bundle)
	if err != nil {
		return err
	}

	issues := bundle.Verify()
	if issues.HasIssues() {
		printIssues(out, issues)
		return issues
	}

	core.Print(out, "%s verified", bundle.Manifest.Name)
	core.Print(out, "hash chain:")
	return printBundleFile(root, parsed.Bundle, bundle.Manifest.Verification.Chain, out)
}

func runInfo(parsed invocation, out io.Writer) error {
	return printBundleFile(bundleRoot(parsed), parsed.Bundle, "manifest.yaml", out)
}

func runLaunch(ctx context.Context, c *core.Core, parsed invocation, out io.Writer) error {
	root := bundleRoot(parsed)
	service := play.NewService(os.DirFS(root), nil)
	plan, err := service.PreparePlay(play.PlayRequest{BundlePath: parsed.Bundle})
	if err != nil {
		return err
	}
	if plan.Issues.HasIssues() {
		printIssues(out, plan.Issues)
		return plan.Issues
	}
	if plan.Engine == nil {
		return core.E("play.launch", "runtime engine is not registered", nil)
	}

	bundle, err := play.LoadBundle(os.DirFS(root), parsed.Bundle)
	if err != nil {
		return err
	}
	home, err := playHome()
	if err != nil {
		return err
	}
	sandbox, err := play.PrepareSandbox(bundle, home, localBundleWriter{})
	if err != nil {
		return err
	}

	return plan.Engine.Run(plan.Manifest.Artefact.Path, play.EngineConfig{
		Core:             c,
		Context:          ctx,
		WorkingDirectory: path.Join(root, parsed.Bundle),
		ConfigPath:       plan.Manifest.Runtime.Config,
		Profile:          plan.Manifest.Runtime.Profile,
		SaveRoot:         sandbox.Root,
		NetworkAllowed:   sandbox.NetworkAllowed,
		Output:           out,
	})
}

func runBundle(parsed invocation, out io.Writer) error {
	artefactData, err := os.ReadFile(parsed.ROM)
	if err != nil {
		return err
	}

	artefactPath := path.Join("rom", path.Base(parsed.ROM))
	service := play.NewService(nil, nil)
	rendered, err := service.RenderBundle(play.BundleRequest{
		Name:           parsed.Name,
		Title:          parsed.Title,
		Author:         parsed.Author,
		Year:           parsed.Year,
		Platform:       parsed.Platform,
		Genre:          parsed.Genre,
		Licence:        parsed.Licence,
		Engine:         parsed.Engine,
		Profile:        parsed.Profile,
		ArtefactPath:   artefactPath,
		ArtefactData:   artefactData,
		ArtefactSHA256: hashBytes(artefactData),
		ArtefactSize:   int64(len(artefactData)),
		ArtefactSource: parsed.Source,
		BYOROM:         parsed.BYOROM,
	})
	if err != nil {
		return err
	}

	targetRoot := bundleRoot(parsed)
	if err := rendered.Write(localBundleWriter{Root: targetRoot}); err != nil {
		return err
	}

	core.Print(out, "bundle created: %s", outputPath(targetRoot, rendered.Path))
	return nil
}

func bundleRoot(parsed invocation) string {
	if parsed.Root != "" {
		return parsed.Root
	}
	if root := os.Getenv("CORE_PLAY_ROOT"); root != "" {
		return root
	}

	return "."
}

func playHome() (string, error) {
	if home := os.Getenv("CORE_PLAY_HOME"); home != "" {
		return home, nil
	}

	return os.UserHomeDir()
}

func printBundleFile(root string, bundlePath string, filePath string, out io.Writer) error {
	data, err := os.ReadFile(path.Join(root, bundlePath, filePath))
	if err != nil {
		return err
	}

	core.Print(out, "%s", string(data))
	return nil
}

func printIssues(out io.Writer, issues play.ValidationErrors) {
	for _, issue := range issues {
		core.Print(out, "%s", issue.Error())
	}
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func defaultText(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

type localBundleWriter struct {
	Root string
}

func (writer localBundleWriter) EnsureDirectory(targetPath string) error {
	return os.MkdirAll(writer.target(targetPath), 0755)
}

func (writer localBundleWriter) WriteFile(targetPath string, data []byte) error {
	return os.WriteFile(writer.target(targetPath), data, 0644)
}

func (writer localBundleWriter) target(targetPath string) string {
	return outputPath(writer.Root, targetPath)
}

func outputPath(root string, targetPath string) string {
	if root == "" || root == "." {
		return targetPath
	}

	return path.Join(root, targetPath)
}
