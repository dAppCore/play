package play

import (
	"testing"

	"dappco.re/go/core"
)

func TestCommand_Commands_Good(testingT *testing.T) {
	testingT.Parallel()

	commands := Commands()
	if len(commands) != 4 {
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
	if len(commands) != 4 {
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

	if len(c.Commands()) != 4 {
		testingT.Fatalf("duplicate Register changed command count: %d", len(c.Commands()))
	}
}
