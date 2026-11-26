package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unyielding Krumar — Creature — Orc Warrior {3}{B}, 3/3 (Khans of Tarkir):
//
//	"{1}{W}: This creature gains first strike until end of turn."
//
// Roofstalker Wight's self grant on its own mana cost: a layer-6 grant
// pinned to the Krumar (CR 613.1f) that lasts until the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00bcfbaf-cd31-4688-973e-4238749051b7",
		Name:         "Unyielding Krumar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Unyielding Krumar — first strike until end of turn"),
		}},
	})
}
