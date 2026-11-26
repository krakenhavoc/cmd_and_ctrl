package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divine Offering — Instant {1}{W} (Mirrodin Besieged, S82 OLD b467, #5576):
//
//	"Destroy target artifact. You gain life equal to its mana value."
//
// The mana value is read off the artifact as it is before it is destroyed
// (CR 608.2h, an {X} in the cost is 0 on the battlefield, CR 202.3e), and the
// life is gained whether or not the destroy worked (indestructible, a
// regeneration shield), as long as the target was still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "231f8edb-4ea1-44be-8794-b76a31462dfc",
		Name:         "Divine Offering",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			mv := c.ManaValue()
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: mv}.Apply(ctx)
		},
	})
}
