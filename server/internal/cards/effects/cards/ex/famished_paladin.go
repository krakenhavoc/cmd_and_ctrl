package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Famished Paladin — Creature — Vampire Knight {1}{W}, 3/3 (Rivals of
// Ixalan, #5531):
//
//	"This creature doesn't untap during your untap step.
//	 Whenever you gain life, untap this creature."
//
// Lurking Roper's text: the untap-step restriction, and an untap on every
// life-gain event of yours (not optional).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "f80df9c4-f682-4a70-81fe-03c7c5c256e1",
		Name:                  "Famished Paladin",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Famished Paladin — untap it", effects.B31UntapSelf),
		},
	})
}
