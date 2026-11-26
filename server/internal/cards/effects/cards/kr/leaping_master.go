package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leaping Master — Creature — Human Monk {1}{R}, 2/1 (Khans of Tarkir):
//
//	"{2}{W}: This creature gains flying until end of turn."
//
// Unyielding Krumar's self grant with flying: a layer-6 grant pinned to the
// creature (CR 613.1f) that lasts until the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b5c2ba0-3df7-4394-a3f1-8e8330d7ac81",
		Name:         "Leaping Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Leaping Master — flying until end of turn"),
		}},
	})
}
