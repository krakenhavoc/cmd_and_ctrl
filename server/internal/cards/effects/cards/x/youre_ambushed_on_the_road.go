package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You're Ambushed on the Road — Instant {W} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Make a Retreat — Return target creature you control to its owner's hand.
//	 • Stand and Fight — Target creature gets +1/+3 until end of turn."
//
// A bounce of the caster's own creature and a pump of any creature, each
// on its bullet's own target judged as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "934ef52e-3e97-418a-bbcf-114badbb2c8a",
		Name:         "You're Ambushed on the Road",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Make a Retreat — Return target creature you control to its owner's hand.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.BounceToHand{Target: id}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("Stand and Fight — Target creature gets +1/+3 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(1, 3, "You're Ambushed on the Road — Stand and Fight")), effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3))),
		),
	})
}
