package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hama Pashar, Ruin Seeker — Legendary Creature — Human Wizard
// {1}{W}{U}, 2/3:
//
//	"Room abilities of dungeons you own trigger an additional time."
//
// A CR 603.2d trigger doubler (#752) whose query matches a triggered
// ability whose source is a dungeon card (its catalog key is a dungeon
// key, game.DungeonKeyPrefix) owned by Hama Pashar's controller. The
// Undercity is a dungeon too, so the initiative's rooms double.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7371040-160c-4502-b7f9-77b7f54cc020",
		Name:         "Hama Pashar, Ruin Seeker",
		Completeness: CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{{
			Label: "Hama Pashar, Ruin Seeker — room abilities trigger an additional time",
			Applies: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
				return strings.HasPrefix(q.Source.OracleID, game.DungeonKeyPrefix) && q.Source.Owner == q.Doubler.Controller
			},
		}},
	})
}
