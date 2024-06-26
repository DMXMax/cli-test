/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DMXMax/cli-test/cmd/scene"
	gdb "github.com/DMXMax/cli-test/util/game"

	"github.com/DMXMax/cli-test/cmd/game"
	gamelog "github.com/DMXMax/cli-test/cmd/log"
	"github.com/DMXMax/cli-test/cmd/roll"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "github.com/DMXMax/cli-test",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Hello, World!")
		cmd.Usage()
		return nil
	},
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "A simple shell, holding token to use interactive commands",
	Long: `A simple shell, holding token to use interactive commands -- the long version
	Examples go here.
`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) error {
		for {
			if gdb.Current != nil {
				g := gdb.Current
				fmt.Print(g.Name, "> ")
			} else {
				fmt.Print("shell> ")
			}

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Split(bufio.ScanLines)
			scanner.Scan()
			//fmt.Print([]string{scanner.Text()})
			//fmt.Print("You entered: ", []string{scanner.Text()}, "\n")
			if scanner.Err() != nil {
				return scanner.Err()
			}
			text := strings.Fields(scanner.Text())
			newCmd, args, err := cmd.Find(text)
			//fmt.Println("newCmd: ", newCmd, "args: ", args, "err: ", err)
			if err != nil {
				cmd.Println(err)
				continue
			}
			if newCmd == cmd {
				continue
			}
			//newCmd.SetArgs(args)

			newCmd.Flags().Parse(args)

			err = newCmd.RunE(newCmd, args)

			cmd.PersistentPostRunE(newCmd, args)

			if err != nil {
				cmd.Println(err)
			}

		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			f.Value.Set("")
			f.Changed = false
		})
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var shellQuitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Quit the shell",
	Long: `Quit the shell, but longer",
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Goodbye!")
		os.Exit(0)
		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli-test.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	shellCmd.AddCommand(shellQuitCmd, scene.SceneCmd, game.GameCmd,
		roll.RollCmd, gamelog.LogCmd, shellHelpCommand)

	rootCmd.AddCommand(shellCmd)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//shellCmd.SetUsageTemplate(Template)
}

var shellHelpCommand = &cobra.Command{
	Use:   "help",
	Short: "Shell help",
	Long:  `Provides detailed shell help help`,
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
