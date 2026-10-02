package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorching Lava — Instant {1}{R}:
//
//	"Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	 Scorching Lava deals 2 damage to any target. If this spell was
//	 kicked, that creature can't be regenerated this turn and if it would
//	 die this turn, exile it instead."
//
// Unkicked it is a Shock. Kicked, the riders are Disintegrate's
// (markCreatureNoRegenExileIfDies): effects of the spell, not of the
// damage, so a creature target is marked whatever it was dealt, and a
// player or a noncreature permanent is only dealt the damage ("that
// creature" names nothing for them).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "e42fb51e-254a-43ac-ad02-9f0fad0f4c8a",
		Name:          "Scorching Lava",
		Completeness:  CompletenessFull,
		Targets:       TargetAny(),
		OptionalCosts: []game.AdditionalCost{Kicker("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return markCreatureNoRegenExileIfDies(ctx, t.ID)
		},
	})
}
