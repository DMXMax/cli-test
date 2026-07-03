package game

import (
	"fmt"
	"math/rand"

	"github.com/DMXMax/mge/util/plot"
	gdb "github.com/DMXMax/mythic-cli/util/game"
	"github.com/spf13/cobra"
)

// plotPointCmd generates a random plot point based on the current game's story themes.
// It rolls on the plot chart and selects a random theme to generate a plot point description.
// Use the --verbose flag to see detailed information about the roll and modifiers.
var plotPointCmd = &cobra.Command{
	Use:     "plotpoint",
	Aliases: []string{"pp", "plot"},
	Short:   "Generate a random plot point for the current game",
	Long:    `Generate a random plot point based on the current game's story themes. Use --verbose to see roll details.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if gdb.Current == nil {
			return fmt.Errorf("no game selected. Use 'game load <name>' to select one")
		}

		roll := rand.Intn(100) + 1
		pickTheme := gdb.Current.StoryThemes.GetRandomTheme() // By default, if no subcommand is given, show help.
		pp, err := plot.Chart.GetChartEntry(roll, pickTheme)
		if err != nil {
			return err
		}

		// The main output is just the description
		cmd.Println(pp.Description)

		// All other information is shown with the verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			cmd.Printf("\n--- Debug Info ---\n")
			cmd.Printf("Theme: %s\tRoll: %d\n", pickTheme, roll)
			cmd.Printf("%-10s %-10s %-10s %-10s %-10s\n", "Action", "Mystery", "Personal", "Social", "Tension")
			cmd.Printf("%-10d %-10d %-10d %-10d %-10d\n", pp.Action, pp.Mystery, pp.Personal, pp.Social, pp.Tension)
		}
		return nil
	},
}

func init() {
	plotPointCmd.Flags().BoolP("verbose", "v", false, "Show verbose output including theme, roll, and modifiers")
}
