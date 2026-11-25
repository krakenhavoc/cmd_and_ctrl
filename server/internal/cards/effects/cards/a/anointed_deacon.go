package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anointed Deacon — Creature — Vampire Cleric {4}{B}, 3/3 (Ixalan):
//
//	"At the beginning of combat on your turn, you may have target Vampire get
//	 +2/+0 until end of turn."
//
// Blightcaster's optional targeting trigger on the beginning of combat: the
// Vampire (any player's, the Deacon included) is chosen as the trigger goes
// on the stack and the "may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	const label = "Anointed Deacon — target Vampire gets +2/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "5212f0a8-7341-4960-835b-2f4600ca1b2d",
		Name:         "Anointed Deacon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Optional(effects.Targeting(
				effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true), label,
					effects.BoostTargetUntilEOT(label, 2, 0)),
				effects.TargetCreature("target Vampire", effects.Subtype("Vampire"))),
				"Anointed Deacon — have target Vampire get +2/+0 until end of turn?"),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0))),
		},
	})
}
