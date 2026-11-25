package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smash to Smithereens — Instant {1}{R} (Magic Origins, S82 older sets):
//
//	"Destroy target artifact. Smash to Smithereens deals 3 damage to that
//	 artifact's controller."
//
// Shatter's destroy with a second sentence that reads the artifact's controller
// as it last was, before it is destroyed (CR 608.2h): the damage is dealt
// whether or not the artifact really died (indestructible, regenerated), since
// the spell does both things it says. A target gone by resolution makes the spell
// do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0583f29-c344-4b56-b0dc-bef1089a95c7",
		Name:         "Smash to Smithereens",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found {
				return nil
			}
			controller := c.Controller
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 3}.Apply(ctx)
		},
	})
}
