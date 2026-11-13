package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crosis's Charm — Instant {U}{B}{R} (Planeshift, #5615):
//
//	"Choose one —
//	 • Return target permanent to its owner's hand.
//	 • Destroy target nonblack creature. It can't be regenerated.
//	 • Destroy target artifact."
//
// A modal spell (CR 700.2) whose bullets are the catalog's mode bodies:
// BounceTheModesTarget to the OWNER's hand, so a stolen permanent goes
// home; Malice's destruction, which a regeneration shield does not stop
// (CR 701.19c); and DestroyTheModesTarget. Each bullet's target is
// checked against its own clause at announce and again at resolution
// (CR 608.2b), so a creature that turned black in response survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e59d70a2-40ac-45b0-995d-65b9caa290c8",
		Name:         "Crosis's Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Return target permanent to its owner's hand.",
				TargetPermanent("target permanent"),
				BounceTheModesTarget),
			ModeDoing("Destroy target nonblack creature. It can't be regenerated.",
				TargetCreature("target nonblack creature", NonBlack()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return DestroyTarget{Target: t.ID, CantBeRegenerated: true}.Apply(ctx)
				}),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
		),
	})
}
