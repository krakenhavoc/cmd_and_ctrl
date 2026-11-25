package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anticognition — Instant {1}{U} (Zendikar Rising, #5508):
//
//	"Counter target creature or planeswalker spell unless its controller pays
//	 {2}. If an opponent has eight or more cards in their graveyard, instead
//	 counter that spell, then scry 2."
//
// The graveyard condition is read as the spell resolves, from Anticognition's
// controller's side (any opponent of theirs). When it holds the spell is
// countered outright (Spell Pierce's tax is not offered) and its caster
// scries two; when it does not the spell is countered unless its controller
// pays {2}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4728e320-d29b-4019-8ab5-e45319c77c40",
		Name:         "Anticognition",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Targets:      effects.TargetSpell("target creature or planeswalker spell", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if !effects.S82OpponentHasGraveyardCards(ctx.Game, ctx.Controller(), 8) {
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{2}",
					Question: "Anticognition — pay {2} or your spell is countered",
				}.Apply(ctx)
			}
			if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.Scry{N: 2}.Apply(ctx)
		},
	})
}
