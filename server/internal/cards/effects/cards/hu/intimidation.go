package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Intimidation — Enchantment {2}{B}{B}{B} (Mercadian Masques, #5619):
//
//	"Creatures you control have fear. (They can't be blocked except by
//	 artifact creatures and/or black creatures.)"
//
// Rage Reflection's team grant, with fear: a layer 6 effect (CR 613.1f) over
// the creatures its controller controls, reaching a creature the moment it
// enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a345d7d1-5d30-49a0-b139-a46f9419c50f",
		Name:         "Intimidation",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16GrantKeywords(effects.B16CreaturesYouControl, "fear")},
	})
}
