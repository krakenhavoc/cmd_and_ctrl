package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashen Monstrosity — Creature — Spirit {5}{R}{R}, 7/4 (Betrayers of
// Kamigawa, #5602):
//
//	"Haste
//	 This creature attacks each combat if able."
//
// Berserkers of Blood Ridge's attack requirement (CR 508.1d): the engine
// refuses to finish declaring attackers while the Monstrosity could attack and
// doesn't. Haste rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01906aae-fc53-40b0-a386-dae43e4f437f",
		Name:            "Ashen Monstrosity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
