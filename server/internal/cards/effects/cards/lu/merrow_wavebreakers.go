package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merrow Wavebreakers — Creature — Merfolk Soldier {4}{U}, 3/3 (Shadowmoor, S82
// OLD b531, #5591):
//
//	"{1}{U}, {Q}: This creature gains flying until end of turn. ({Q} is the untap
//	 symbol.)"
//
// {Q} is AbilityCost.UntapSelf (CR 107.6), so the Wavebreakers is tapped to pay it
// and untaps as the cost: it can then attack with flying. The keyword ends at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6c325e8-b9bd-44ab-9370-196a40c71e1f",
		Name:         "Merrow Wavebreakers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {Q}: This creature gains flying until end of turn.",
			Cost:    effects.OldB531UntapCost("{1}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Effect:  effects.OldB531ThisGainsUntilEOT("Merrow Wavebreakers — flying until end of turn", "flying"),
		}},
	})
}
