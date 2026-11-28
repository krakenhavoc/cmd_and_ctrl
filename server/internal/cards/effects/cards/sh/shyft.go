package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shyft — Creature — Shapeshifter {4}{U}, 4/2 (Ice Age, S82 OLD, #5636):
//
//	"At the beginning of your upkeep, you may have this creature become the color or colors of your
//	 choice. (This effect lasts indefinitely.)"
//
// The "may" is asked as the trigger resolves, then one or more colors are chosen (Scuttlemutt's color
// prompt). The colors replace the old ones for as long as it stays this object (CR 611.2a, 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40c3007a-be6c-4400-9139-12b9ca098ac4",
		Name:         "Shyft",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Shyft — you may have it become the color or colors of your choice", effects.S82OldB843ShyftUpkeep),
		},
	})
}
