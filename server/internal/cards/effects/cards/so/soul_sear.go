package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1ef73b3b-ab4a-494b-b969-dcde995cee34",
		Name:         "Soul Sear",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				ctx.RecordStep(soulSearRecord, []uuid.UUID{id})
				return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if ids := ctx.StepRecord(soulSearRecord).IDs; len(ids) > 0 {
					return effects.G2LosesIndestructibleUntilEOT(ctx, ids[0])
				}
				return nil
			}),
		),
	})
}

// soulSearRecord is the permanent that was dealt damage.
const soulSearRecord = "soul-sear/target"
