package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunken Hope — Enchantment {3}{U}{U} (Planechase 2012, S82 OLD, #5569):
//
//	"At the beginning of each player's upkeep, that player returns a creature
//	 they control to its owner's hand."
//
// Ruthless Winnower's every-upkeep edict, with a bounce for the sacrifice: the
// player whose upkeep it is chooses one creature they control (not asked with
// exactly one, nothing with none) and returns it to its owner's hand. The choice
// is theirs, not the Hope's controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1820b6e9-83fd-40b9-953e-ba1a8ae69614",
		Name:         "Sunken Hope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Sunken Hope — that player returns a creature they control to its owner's hand", effects.S82OldB432SunkenHopeBounce),
		},
	})
}
