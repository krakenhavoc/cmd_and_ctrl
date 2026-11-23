package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call to Heel — Instant {1}{U} (Shards of Alara, #5588):
//
//	"Return target creature to its owner's hand. Its controller draws a
//	 card."
//
// Dream Fracture's "its controller draws" on a bounce: the controller is
// read before the creature leaves (CR 608.2h), so a stolen creature goes
// to its owner's hand and the player who controlled it draws. The draw is
// not contingent on the return: a creature a replacement sends elsewhere
// still has a controller who draws. An illegal target makes the spell do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ef3dd68-f6a1-43b9-8837-40b326019b1b",
		Name:         "Call to Heel",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			creature, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: creature.Controller, N: 1}.Apply(ctx)
		},
	})
}
