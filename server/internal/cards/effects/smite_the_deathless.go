package effects

import "github.com/google/uuid"

// Smite the Deathless — Instant {1}{R}:
//
//	"Smite the Deathless deals 3 damage to target creature. That creature
//	 loses indestructible until end of turn. If that creature would die
//	 this turn, exile it instead."
//
// Both riders are the spell's, on the target whether or not the damage
// was dealt (ADR 0108 §1). In printed order: the damage, then the
// layer-6 loss of indestructible (CR 613.1f), then the replacement. The
// state-based check after resolution then destroys an indestructible
// creature carrying lethal damage (CR 702.12b, 704.5g), and it is
// exiled.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "37994591-3494-4314-a7da-49c112b0866f",
		Name:         "Smite the Deathless",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Targets:      TargetCreature("target creature"),
		// Each sentence is a step: the creature loses indestructible and is exiled
		// if it would die after a prompt the damage asks (a CR 616 order; CR
		// 608.2c, #4155). The target is read once, as the spell resolves.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				ctx.RecordStep(smiteTheDeathlessRecord, []uuid.UUID{id})
				return DealDamage{Source: ctx.Source(), Target: id, Amount: 3}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if ids := ctx.StepRecord(smiteTheDeathlessRecord).IDs; len(ids) > 0 {
					return g2LosesIndestructibleUntilEOT(ctx, ids[0])
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if ids := ctx.StepRecord(smiteTheDeathlessRecord).IDs; len(ids) > 0 {
					return ExileIfItWouldDieThisTurn{Target: ids[0]}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}

// smiteTheDeathlessRecord is the creature that was dealt damage.
const smiteTheDeathlessRecord = "smite-the-deathless/target"
