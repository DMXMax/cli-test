/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
// Package cmd provides the command-line interface for the Mythic CLI application.
// It implements the root command, interactive shell, and command routing.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DMXMax/mge/chart"
	"github.com/DMXMax/mythic-cli/cmd/descriptor"
	"github.com/DMXMax/mythic-cli/cmd/scene"
	gdb "github.com/DMXMax/mythic-cli/util/game"

	"github.com/DMXMax/mythic-cli/cmd/game"
	gamelog "github.com/DMXMax/mythic-cli/cmd/log"
	"github.com/DMXMax/mythic-cli/cmd/roll"

	"github.com/DMXMax/mythic-cli/util/input"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rootCmd represents the base command when called without any subcommands.
// It displays usage information when invoked directly.
var rootCmd = &cobra.Command{
	Use:   "mythic-cli",
	Short: "Mythic Game Master Emulator CLI",
	Long: `A command-line tool for solo RPG gaming using the Mythic Game Master Emulator system.

Features:
- Interactive shell for game management
- Dice rolling with Mythic fate chart
- Game state persistence with SQLite
- Story logging and chaos factor management
- Character and scene management

Perfect for solo RPG adventures, GM-less gaming, and story generation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Usage()
		return nil
	},
}

// errQuit is a sentinel error to signal a clean exit from the shell.
// When returned from a command, it causes the shell loop to exit gracefully.
var errQuit = errors.New("user requested quit")

// shellCmd provides an interactive shell for running Mythic CLI commands.
// It supports command history, line editing, and a dynamic prompt that shows
// the current game state (name and chaos factor).
var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start an interactive shell for game management",
	Long: `Start an interactive shell that allows you to run Mythic CLI commands
with persistent history and line editing support.

The shell prompt displays the current game name and chaos factor (C)
when a game is loaded. Use 'quit' or press Ctrl-C/Ctrl-D to exit.

Command history is persisted to ~/.mythic-cli_history and can be navigated
using the Up/Down arrow keys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use liner to get arrow-key history and line editing
		l := liner.NewLiner()
		defer l.Close()
		l.SetCtrlCAborts(true)
		// Make liner available to commands for sub-prompts
		input.SetPrompter(l)

		// Load/save persistent history
		home, _ := os.UserHomeDir()
		histPath := filepath.Join(home, ".mythic-cli_history")
		if f, err := os.Open(histPath); err == nil {
			_, _ = l.ReadHistory(f)
			_ = f.Close()
		}
		defer func() {
			if f, err := os.Create(histPath); err == nil {
				_, _ = l.WriteHistory(f)
				_ = f.Close()
			}
		}()

		for {
			var prompt string
			if gdb.Current != nil {
				g := gdb.Current
				prompt = fmt.Sprintf("%s (C:%d)> ", g.Name, chart.ChaosInternalToUser(int(g.Chaos)))
			} else {
				prompt = "shell> "
			}

			line, err := l.Prompt(prompt)
			if err == liner.ErrPromptAborted { // Ctrl-C
				cmd.Println("Goodbye!")
				return nil
			}
			if err == io.EOF { // Ctrl-D
				cmd.Println("Goodbye!")
				return nil
			}
			if err != nil {
				return err
			}

			input := strings.TrimSpace(line)
			if input == "" {
				continue
			}

			fields := strings.Fields(input)
			var newCmd *cobra.Command
			var newArgs []string
			var findErr error

			// Shell-only commands live under shellCmd; everything else on rootCmd.
			if fields[0] == "quit" || fields[0] == "help" {
				newCmd, newArgs, findErr = cmd.Find(fields)
			} else {
				newCmd, newArgs, findErr = rootCmd.Find(fields)
			}
			if findErr != nil {
				cmd.Println(findErr)
				continue
			}
			if newCmd == cmd || newCmd == rootCmd {
				cmd.Printf("Error: unknown command \"%s\" for \"%s\"\n", fields[0], cmd.CommandPath())
				continue
			}

			// Append to history before execution
			l.AppendHistory(input)

			// Check if help is requested
			hasHelp := false
			for _, arg := range newArgs {
				if arg == "--help" || arg == "-h" {
					hasHelp = true
					break
				}
			}

			if hasHelp {
				newCmd.Help()
			} else {
				newCmd.SetOut(cmd.OutOrStdout())
				newCmd.SetErr(cmd.ErrOrStderr())
				newCmd.SetArgs(newArgs)
				if err := newCmd.Execute(); err != nil {
					if errors.Is(err, errQuit) {
						return nil
					}
					cmd.Println(err)
				}

				newCmd.Flags().VisitAll(func(f *pflag.Flag) {
					f.Value.Set(f.DefValue)
					f.Changed = false
				})
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// If execution fails, the program exits with code 1.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// shellQuitCmd allows the user to gracefully exit the interactive shell.
var shellQuitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Quit the interactive shell",
	Long:  `Exit the interactive shell. You can also use Ctrl-C or Ctrl-D to exit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Goodbye!")
		return errQuit
	},
}

func init() {
	shared := []*cobra.Command{
		scene.SceneCmd, game.GameCmd, roll.RollCmd, roll.RollFateCmd,
		gamelog.LogCmd, descriptor.DescriptorCmd,
	}
	for _, c := range shared {
		rootCmd.AddCommand(c)
	}

	shellCmd.AddCommand(shellQuitCmd, shellHelpCommand)

	rootCmd.AddCommand(shellCmd)

	// Root command flags (currently unused, but available for future use)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// shellHelpCommand provides help functionality within the interactive shell.
// It allows users to get help for any command by typing "help <command>".
var shellHelpCommand = &cobra.Command{
	Use:   "help",
	Short: "Show help for commands",
	Long:  `Show help information for commands. Use "help <command>" to get detailed help for a specific command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		newcmd, _, err := cmd.Parent().Find(args)
		if err != nil {
			return err
		}
		if newcmd == cmd {
			return cmd.Parent().Usage()
		}
		return newcmd.Help()
	},
}
