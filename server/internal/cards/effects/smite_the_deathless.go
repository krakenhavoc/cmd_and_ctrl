package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if err := g2LosesIndestructibleUntilEOT(ctx, id); err != nil {
				return err
			}
			return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
		},
	})
}
