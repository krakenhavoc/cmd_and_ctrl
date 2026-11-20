package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:              "f80df9c4-f682-4a70-81fe-03c7c5c256e1",
		Name:                  "Famished Paladin",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Famished Paladin — untap it", b31UntapSelf),
		},
	})
}
