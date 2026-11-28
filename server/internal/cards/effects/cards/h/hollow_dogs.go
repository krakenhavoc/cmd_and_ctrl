package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hollow Dogs — {4}{B} Creature — Phyrexian Zombie Dog 3/3 (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Mortis Dogs' attack trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16aa2e9e-bfe8-4b40-a1db-bb1da2ede849",
		Name:         "Hollow Dogs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Hollow Dogs — it gets +2/+0 until end of turn",
				effects.ThisCreatureUntilEOT("Hollow Dogs — +2/+0", 2, 0)),
		},
	})
}
