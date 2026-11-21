package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voracious Greatshark — Creature — Shark {3}{U}{U}, 5/4 (Foundations,
// #4465):
//
//	"Flash
//	 When this creature enters, counter target artifact or creature
//	 spell."
//
// Flash from PrintedKeywords, and an entry trigger with Hurl into
// History's spell target. The spell is chosen as the trigger goes on
// the stack and countered if it is still a legal target as the trigger
// resolves (CR 608.2b); with no artifact or creature spell on the
// stack, the trigger is removed for want of a target (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d96840e-f9b3-4c86-8f59-f12e5736318f",
		Name:            "Voracious Greatshark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Voracious Greatshark — counter target artifact or creature spell", effects.CounterTheChosenSpell),
				effects.TargetSpell("target artifact or creature spell", effects.Or(effects.Artifact(), effects.Creature()))),
		},
	})
}
