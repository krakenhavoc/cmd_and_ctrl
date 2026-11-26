package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grisly Spectacle — Instant {2}{B}{B} (Gatecrash, #5565):
//
//	"Destroy target nonartifact creature. Its controller mills cards equal to
//	 that creature's power."
//
// The creature's controller and power are read before it is destroyed
// (CR 608.2h, last-known information), and its controller mills that many
// cards whether or not the destruction happened (an indestructible creature
// still costs its controller the cards). A power of 0 or less mills nothing
// (CR 107.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c83d21a-6138-409f-8e00-75dcd14388e2",
		Name:         "Grisly Spectacle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonartifact creature", effects.Not(effects.Artifact())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				controller, power := c.Controller, effects.DepartedCreaturePower(ctx.Game, t.ID)
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.MillCards{Player: controller, N: power}.Apply(ctx)
			}
			return nil
		},
	})
}
