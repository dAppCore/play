package main

import (
	"context"
	"flag"
	"io"
	"os"
	"path"

	"dappco.re/go/core"
	"dappco.re/go/play"
)

type operation string

const (
	operationLaunch       operation = "launch"
	operationList         operation = "list"
	operationVerify       operation = "verify"
	operationShieldVerify operation = "shield-verify"
	operationInfo         operation = "info"
	operationBundle       operation = "bundle"
	operationEngines      operation = "engines"
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
	EngineBin string
	Profile   string
	ROM       string
	Source    string
	CPU       int
	Memory    int64
	BYOROM    bool
	JSON      bool
	Archive   bool
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
	case operationVerify, operationShieldVerify:
		return runVerify(parsed, out)
	case operationInfo:
		return runInfo(parsed, out)
	case operationBundle:
		return runBundle(parsed, out)
	case operationEngines:
		return runEngines(out)
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
	case "shield-verify":
		return parseNamed(operationShieldVerify, args[1:])
	case "info":
		return parseNamed(operationInfo, args[1:])
	case "engines":
		return invocation{Operation: operationEngines}, nil
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
	case "list":
		return parseList(args[1:])
	case "--list":
		return parseList(args[1:])
	case "verify":
		return parseNamed(operationVerify, args[1:])
	case "--verify":
		return parseNamed(operationVerify, args[1:])
	case "shield-verify":
		return parseNamed(operationShieldVerify, args[1:])
	case "info":
		return parseNamed(operationInfo, args[1:])
	case "--info":
		return parseNamed(operationInfo, args[1:])
	case "engines":
		return invocation{Operation: operationEngines}, nil
	case "bundle":
		return parseBundle(args[1:])
	default:
		return parseLaunch(args)
	}
}

func parseList(args []string) (invocation, error) {
	flags := newFlagSet("list")
	root := flags.String("root", "", "bundle root")
	jsonOutput := flags.Bool("json", false, "JSON output")
	if err := flags.Parse(args); err != nil {
		return invocation{}, err
	}
	if flags.NArg() > 0 {
		return invocation{}, core.E("play.parse", "list does not accept a bundle argument", nil)
	}

	return invocation{Operation: operationList, Root: *root, JSON: *jsonOutput}, nil
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
	engineBinary := flags.String("engine-binary", "", "engine binary path")
	profile := flags.String("profile", "", "runtime profile")
	rom := flags.String("rom", "", "artefact path")
	source := flags.String("source", "", "artefact source")
	cpu := flags.Int("cpu-percent", 0, "CPU limit percentage")
	memory := flags.Int64("memory-bytes", 0, "memory limit in bytes")
	byorom := flags.Bool("byorom", false, "BYOROM mode")
	archive := flags.Bool("archive", false, "write deterministic zip archive")
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
		EngineBin: *engineBinary,
		Profile:   *profile,
		ROM:       *rom,
		Source:    *source,
		CPU:       *cpu,
		Memory:    *memory,
		BYOROM:    *byorom,
		Archive:   *archive,
	}, nil
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
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
	if plan.Launch != nil {
		sandboxIssues := sandbox.ValidateLaunch(*plan.Launch)
		if sandboxIssues.HasIssues() {
			printIssues(out, sandboxIssues)
			return sandboxIssues
		}
	}

	return plan.Engine.Run(plan.Manifest.Artefact.Path, play.EngineConfig{
		Core:             c,
		Context:          ctx,
		WorkingDirectory: path.Join(root, parsed.Bundle),
		ConfigPath:       plan.Manifest.Runtime.Config,
		Profile:          plan.Manifest.Runtime.Profile,
		SaveRoot:         sandbox.Root,
		Resources:        sandbox.Resources,
		NetworkAllowed:   sandbox.NetworkAllowed,
		Output:           out,
	})
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

func defaultText(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
