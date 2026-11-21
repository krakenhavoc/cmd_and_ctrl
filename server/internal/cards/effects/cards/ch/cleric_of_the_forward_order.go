package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleric of the Forward Order — Creature — Human Cleric {1}{W}, 2/2
// (Dominaria Remastered, #5948):
//
//	"When this creature enters, you gain 2 life for each creature you
//	 control named Cleric of the Forward Order."
//
// Counted as the trigger resolves, so the Cleric that entered counts itself
// (it is one it controls), and two entering together each count both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84a2a592-0e31-434f-b3b4-f2013d86f29d",
		Name:         "Cleric of the Forward Order",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cleric of the Forward Order — you gain 2 life for each creature you control named Cleric of the Forward Order",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
						return c.IsCreature() && c.Name == "Cleric of the Forward Order"
					})
					if n <= 0 {
						return nil
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 2 * n}.Apply(ctx)
				}))),
		},
	})
}
