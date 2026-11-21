package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voldaren Duelist — Creature — Vampire Warrior {3}{R}, 3/2 (Foundations
// Jumpstart, S82, #5929):
//
//	"Haste
//	 When this creature enters, target creature can't block this turn."
//
// Hammerhand's enters trigger (any creature is a legal target).
//
// No simplification.
func init() {
	const label = "Voldaren Duelist — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:        "a96d0496-babc-4774-9b72-bea9d2b22743",
		Name:            "Voldaren Duelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
					effects.TargetCreature("target creature")),
				effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
