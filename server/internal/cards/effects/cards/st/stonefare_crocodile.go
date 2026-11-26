package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonefare Crocodile — Creature — Crocodile {2}{G}, 3/2 (Return to Ravnica,
// #5566):
//
//	"{2}{B}: This creature gains lifelink until end of turn."
//
// The grant is to this permanent only while it is still the one that activated
// the ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fef48e69-ab4d-4364-be41-7560440afc18",
		Name:         "Stonefare Crocodile",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: This creature gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("lifelink", "Stonefare Crocodile — lifelink until end of turn"),
		}},
	})
}
