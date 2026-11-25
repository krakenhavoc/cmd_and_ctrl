package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geist of the Lonely Vigil — Creature — Spirit Cleric {1}{W}, 2/3 (Eldritch
// Moon, S82 OLD, #5541):
//
//	"Defender, flying
//	 Delirium — This creature can attack as though it didn't have defender
//	 as long as there are four or more card types among cards in your
//	 graveyard."
//
// Backstreet Bruiser's defender waiver (CR 702.3b), gated on the
// controller's own delirium and read live at declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e80481e-06ed-425a-8991-0557e49aba6a",
		Name:            "Geist of the Lonely Vigil",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"Delirium — This creature can attack as though it didn't have defender as long as there are four or more card types among cards in your graveyard.",
			func(g *game.Game, source *game.Card) bool {
				return effects.DrcDeliriumMet(g, source.Controller)
			})},
	})
}
