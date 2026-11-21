package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mayael's Aria — Enchantment {R}{G}{W} (S82 GAP, #5948):
//
//	"At the beginning of your upkeep, put a +1/+1 counter on each creature
//	 you control if you control a creature with power 5 or greater. Then you
//	 gain 10 life if you control a creature with power 10 or greater. Then
//	 you win the game if you control a creature with power 20 or greater."
//
// Three sentences, each one's "if" read as that sentence is reached, so the
// counters from the first count toward the second and third (a 9-power
// creature becomes a 10-power one and gains the life).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cb9ca93-e0bb-4758-bf0d-c9a8fce5f11f",
		Name:         "Mayael's Aria",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Mayael's Aria — counters, life, then the win",
				effects.Do(effects.Instruction(mayaelsAriaResolve))),
		},
	})
}

func mayaelsAriaResolve(ctx *effects.Context) error {
	g, me := ctx.Game, ctx.Controller()
	if effects.B24ControlsCreatureWithPowerAtLeast(g, me, 5) {
		if err := effects.B11PutCounterOnEachCreatureYouControl(g, ctx.Item); err != nil {
			return err
		}
	}
	g.RecomputeLayersIfStaleLocked()
	if effects.B24ControlsCreatureWithPowerAtLeast(g, me, 10) {
		if err := (effects.GainLife{Amount: 10}).Apply(ctx); err != nil {
			return err
		}
	}
	if effects.B24ControlsCreatureWithPowerAtLeast(g, me, 20) {
		return effects.WinTheGame{Player: me}.Apply(ctx)
	}
	return nil
}
