package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Getaway Glamer — Instant {W} (Outlaws of Thunder Junction):
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Exile target nontoken creature. Return it to the battlefield
//	   under its owner's control at the beginning of the next end step.
//	 + {2} — Destroy target creature if no other creature has greater power."
//
// The first bullet is Abuelo's delayed flicker (the exile now, one delayed
// trigger that returns the card as a new object, CR 400.7). The second
// reads power as it resolves: the target is destroyed only if no OTHER
// creature on the battlefield, any controller's, has greater power, so a tie
// does not save it. With both bullets on the same creature, the first runs
// first (CR 608.2c) and the creature is gone from the battlefield, so the
// second finds no legal target and does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98eac0f7-3f4d-46ef-9575-680c52e967cf",
		Name:         "Getaway Glamer",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Exile target nontoken creature. Return it to the battlefield under its owner's control at the beginning of the next end step.", "{1}",
				TargetCreature("target nontoken creature", Not(IsTokenPredicate())),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return exileCardsThenScheduleReturn(ctx, []uuid.UUID{t.ID}, "Getaway Glamer — return the exiled creature")
				}),
			SpreeModeDoing("Destroy target creature if no other creature has greater power.", "{2}",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					target, found := ctx.Game.LookupCardForEffect(t.ID)
					if !found {
						return nil
					}
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.InstanceID != t.ID && c.IsCreature() && c.CurrentPower() > target.CurrentPower() {
							return nil
						}
					}
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}),
		),
	})
}
