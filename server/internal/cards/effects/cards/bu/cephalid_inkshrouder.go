package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cephalid Inkshrouder — Creature — Octopus {2}{U}, 2/1 (Judgment, #5611):
//
//	"Discard a card: This creature gains shroud until end of turn and can't be
//	 blocked this turn."
//
// One ability, two effects on the same permanent: the shroud keyword and the
// "can't be blocked" restriction, both until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3add0d7-2610-4b4e-9d43-3f04a5cd8259",
		Name:         "Cephalid Inkshrouder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: This creature gains shroud until end of turn and can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.DiscardACard(),
			Effect:  effects.OldB624ShroudAndCantBeBlocked("Cephalid Inkshrouder"),
		}},
	})
}
