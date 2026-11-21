package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurking Roper — Creature — Horror {2}{G}, 4/5 (Adventures in the
// Forgotten Realms, S82):
//
//	"This creature doesn't untap during your untap step.
//	 Whenever you gain life, untap this creature."
//
// Battered Golem's untap-step restriction with an untap on every life-gain
// event (not optional).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "61f3e964-6bea-4543-992e-3e36404bf628",
		Name:                  "Lurking Roper",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Lurking Roper — untap it", effects.B31UntapSelf),
		},
	})
}
