package play

import "testing"

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
