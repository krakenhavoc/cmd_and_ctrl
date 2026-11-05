package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plasma Cascade — Instant {2}{R} (Star Trek, #4973):
//
//	"Plasma Cascade deals 3 damage to any target. You may discard a
//	 card. If you do, draw a card."
//
// The damage comes first; then the caster is asked whether to discard,
// and draws only if a card was discarded (b39MayDiscardThenDraw, the
// Improvised Act shape). A target that has become illegal makes the
// whole spell do nothing (CR 608.2b), so there is no loot either.
//
// No simplification.
func init() {
	const question = "Plasma Cascade — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "46ce40e9-b7fb-4f40-a7ab-f8a2234d1d63",
		Name:         "Plasma Cascade",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			return b39MayDiscardThenDraw(1, false, question, drawAsManyAsDiscarded)(ctx)
		},
	})
}
