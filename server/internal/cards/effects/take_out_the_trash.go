package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Take Out the Trash — Instant {1}{R} (Bloomburrow, #4467):
//
//	"Take Out the Trash deals 3 damage to target creature or
//	 planeswalker. If you control a Raccoon, you may discard a card. If
//	 you do, draw a card."
//
// Plasma Cascade's shape: the damage, then Flaring Cinder's "you may
// discard a card. If you do, draw a card" (b39MayDiscardThenDraw), asked
// only if you control a Raccoon as the spell resolves, after the damage.
// "Raccoon" is read through the spell's text (CR 612.2). A spell whose
// only target has become illegal does not resolve (CR 608.2b), so it
// offers no rummage either.
//
// No simplification.
func init() {
	const question = "Take Out the Trash — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "e07700e7-218a-45a3-9ee4-de2cc25cdc02",
		Name:         "Take Out the Trash",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if !ControlsA(ctx.TextSubtype("Raccoon"))(ctx.Game, ctx.Controller()) {
				return nil
			}
			return b39MayDiscardThenDraw(1, false, question, drawOneIfDiscarded)(ctx)
		},
	})
}
