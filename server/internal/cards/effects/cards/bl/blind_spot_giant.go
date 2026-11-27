package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blind-Spot Giant — Creature — Giant Warrior {2}{R}, 4/3 (Lorwyn, S82 OLD,
// #5593):
//
//	"This creature can't attack or block unless you control another Giant."
//
// Warden of the Chained's attack restriction (read live at declaration over the
// controller's other creatures, CR 508.1c) and Felhide Brawler's block rule
// (CR 509.1b), both over another Giant: the Giant's own type never satisfies it
// (CR 109.1) and a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37165859-87bc-4d44-95a2-39051ab20a14",
		Name:         "Blind-Spot Giant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.CantAttackUnlessYouControlAnother(game.PermanentQuery{Types: []string{"creature"}, Subtypes: []string{"Giant"}}),
				"Can't attack unless you control another Giant."),
		},
		BlockRules: []game.BlockRule{
			effects.CantBlockUnlessYouControlAnother("Giant", "it can't block unless its controller controls another Giant"),
		},
	})
}
