package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dross Hopper — Creature — Phyrexian Insect Horror {1}{B}, 2/1:
//
//	"Sacrifice a creature: This creature gains flying until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ec8c540-73bf-4412-a53a-499e9b9fd213",
		Name:         "Dross Hopper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Dross Hopper — flying until end of turn"),
		}},
	})
}
