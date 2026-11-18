package effects

import "github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "934ef52e-3e97-418a-bbcf-114badbb2c8a",
		Name:         "You're Ambushed on the Road",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Make a Retreat — Return target creature you control to its owner's hand.",
				TargetCreature("target creature you control", YouControl()),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return BounceToHand{Target: id}.Apply(ctx)
				})),
			ModeWithPurpose(ModeDoing("Stand and Fight — Target creature gets +1/+3 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(1, 3, "You're Ambushed on the Road — Stand and Fight")), ForTargets(BenefitUntilEOT(0, 1, 3))),
		),
	})
}
