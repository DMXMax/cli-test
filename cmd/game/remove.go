package game

import (
	"fmt"
	"strings"

	"github.com/DMXMax/mge/storage"
	"github.com/DMXMax/mythic-cli/util/db"
	gdb "github.com/DMXMax/mythic-cli/util/game"
	"github.com/spf13/cobra"
)

var removeName string

// removeCmd permanently removes a game and all its associated log entries from the database.
// If the removed game is currently selected, the current game is cleared.
// The game name can be provided as a positional argument or via the --name flag.
var removeCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm", "delete", "del"},
	Short:   "Remove a game and all its logs",
	Long:    `Remove a game by name. This also removes all associated log entries. You can pass the name as a positional argument or via --name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			// Join all args to handle multi-word names (e.g., "Kat in Shadow")
			name = strings.Join(args, " ")
		} else {
			name = removeName
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("no game name specified")
		}
		name = storage.SanitizeGameName(name)
		if err := storage.ValidateGameName(name); err != nil {
			return err
		}

		// Find the game by name
		var game gdb.Game
		if err := db.GamesDB.Where("name = ?", name).First(&game).Error; err != nil {
			return fmt.Errorf("could not find game '%s': %w", name, err)
		}

		// Permanently delete game and all associated records
		if err := storage.DeleteGamePermanently(db.GamesDB, game.ID); err != nil {
			return fmt.Errorf("failed to delete game '%s': %w", name, err)
		}

		// Clear current selection if it was this game
		if gdb.Current != nil && gdb.Current.ID == game.ID {
			gdb.Current = nil
		}

		cmd.Printf("Removed game: %s (and all logs, scenes, threads, and characters)\n", name)
		return nil
	},
}

func init() {
	removeCmd.Flags().StringVar(&removeName, "name", "", "name of the game to remove")
}
