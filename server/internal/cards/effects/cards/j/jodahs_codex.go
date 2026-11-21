package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jodah's Codex — Artifact — Book {5}:
//
//	"Domain — {5}, {T}: Draw a card. This ability costs {1} less to
//	 activate for each basic land type among lands you control."
//
// The ability's own cost clause (ActivatedAbility.CostModifiers) priced
// by the shared domain count (DomainCostReduction, #5666). Generic mana
// only (CR 601.2f), so with all five types the {5} is free but the {T}
// is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ba5b108-59cb-4af1-88cb-d3ceaaf00784",
		Name:         "Jodah's Codex",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Domain — {5}, {T}: Draw a card. This ability costs {1} less to activate for each basic land type among lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.DomainCostReduction("This ability costs {1} less to activate for each basic land type among lands you control."),
			},
			Effect: effects.B36DrawOne,
		}},
	})
}
