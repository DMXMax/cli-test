package scene

import (
	"fmt"
	"strings"

	"github.com/DMXMax/mge/storage"
	"github.com/DMXMax/mge/util"
	"github.com/DMXMax/mge/util/scene"
	"github.com/DMXMax/mythic-cli/util/db"
	gdb "github.com/DMXMax/mythic-cli/util/game"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// startCmd starts a new scene with an Expected Scene concept.
// It automatically rolls the Chaos Die to determine if the scene is Expected, Altered, or Interrupted.
var startCmd = &cobra.Command{
	Use:   "start <description>",
	Short: "Start a new scene",
	Long: `Start a new scene with an Expected Scene concept. The Chaos Die is automatically rolled
to determine if the scene proceeds as expected, is altered, or is interrupted.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		g := gdb.Current
		if g == nil {
			return fmt.Errorf("no game selected. Use 'game load <name>' to select one")
		}

		concept := strings.Join(args, " ")
		rollResult := scene.RollChaosDie(int(g.Chaos))

		var logMsg string
		var eventDisplay string
		if rollResult.SceneType == "altered" || rollResult.SceneType == "interrupt" {
			event := util.GetEvent()
			eventDisplay = event.String()
			logMsg = fmt.Sprintf("--- Scene Start: %s | Expected: %s | Event: %s ---",
				strings.Title(rollResult.SceneType), concept, eventDisplay)
		} else {
			logMsg = fmt.Sprintf("--- Scene Start: Expected | %s ---", concept)
		}

		err := db.GamesDB.Transaction(func(tx *gorm.DB) error {
			if _, err := storage.StartScene(tx, g.ID, concept, rollResult); err != nil {
				return err
			}
			entry := gdb.LogEntry{
				Type:   0,
				Msg:    logMsg,
				GameID: g.ID,
			}
			return tx.Create(&entry).Error
		})
		if err != nil {
			return fmt.Errorf("failed to start scene: %w", err)
		}

		cmd.Printf("Scene Started: %s\n", rollResult.Description)
		cmd.Printf("Expected Scene: %s\n", concept)
		if eventDisplay != "" {
			cmd.Printf("\nRandom Event: %s\n", eventDisplay)
		}

		return nil
	},
}

func init() {
	SceneCmd.AddCommand(startCmd)
}
