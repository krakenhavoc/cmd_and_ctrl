package effects

import "github.com/google/uuid"

// Soul Sear — Instant {2}{R} (Foundations Jumpstart):
//
//	"Soul Sear deals 5 damage to target creature or planeswalker. That
//	 permanent loses indestructible until end of turn."
//
// Smite the Deathless's shape: the damage, then the layer-6 loss of
// indestructible (CR 613.1f), so the state-based check after resolution
// destroys an indestructible creature carrying lethal damage (CR 702.12b,
// 704.5g). The rider applies whether or not the damage was dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ef73b3b-ab4a-494b-b969-dcde995cee34",
		Name:         "Soul Sear",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				ctx.RecordStep(soulSearRecord, []uuid.UUID{id})
				return DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if ids := ctx.StepRecord(soulSearRecord).IDs; len(ids) > 0 {
					return g2LosesIndestructibleUntilEOT(ctx, ids[0])
				}
				return nil
			}),
		),
	})
}

// soulSearRecord is the permanent that was dealt damage.
const soulSearRecord = "soul-sear/target"
