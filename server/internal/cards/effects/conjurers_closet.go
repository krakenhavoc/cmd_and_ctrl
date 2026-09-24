package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conjurer's Closet — Artifact {5}:
//
//	"At the beginning of your end step, you may exile target creature
//	 you control, then return that card to the battlefield under your
//	 control."
//
// Thassa, Deep-Dwelling's end-step blink with the printed clauses this
// card doesn't share removed: no God-clause static, no activated tap
// ability, and the target is a plain "target creature you control"
// rather than an "up to one OTHER" — Conjurer's Closet has no name of
// its own to avoid, so nothing needs excluding, and the trigger is a
// real "you may" rather than an implicit zero floor.
//
// The exile and the return happen together as one Flicker: the
// creature is a new object (CR 400.7) that re-fires every one of its
// ETBs, which is the whole reason to play the artifact over a plain
// blink spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd1eda60-53e4-44d0-9b2c-7a57395e291f",
		Name:         "Conjurer's Closet",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{conjurersClosetEndStepBlink()},
	})
}

// conjurersClosetEndStepBlink is "At the beginning of your end step,
// you may exile target creature you control, then return that card to
// the battlefield under your control."
func conjurersClosetEndStepBlink() game.TriggeredAbility {
	return Optional(
		Targeting(
			AtYourEndStep("Conjurer's Closet — exile a creature you control, then return it to the battlefield",
				flickerFirstLegalTarget),
			TargetCreature("target creature you control", YouControl())),
		"Conjurer's Closet — exile a creature you control, then return it to the battlefield?")
}
