package effects

import "github.com/google/uuid"

// Compleat Devotion — Instant {1}{W} (Phyrexia: All Will Be One, #5479):
//
//	"Target creature you control gets +2/+2 until end of turn. If that
//	 creature has toxic, draw a card."
//
// The draw is conditional on the creature as the spell resolves, after
// the pump. A target that is gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2a6cfb7-e5bd-49c8-aa51-fd5203eca8b8",
		Name:         "Compleat Devotion",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Compleat Devotion — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if !s82OneB05HasToxic(ctx.Game, id) {
					return nil
				}
				return DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
