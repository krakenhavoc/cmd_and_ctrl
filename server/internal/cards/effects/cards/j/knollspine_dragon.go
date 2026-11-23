package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knollspine Dragon — Creature — Dragon {5}{R}{R}, 7/5 (Tales of
// Middle-earth Commander):
//
//	"Flying
//	 When this creature enters, you may discard your hand and draw cards
//	 equal to the damage dealt to target opponent this turn."
//
// Balin, Loremaster's "may discard your hand" shape with a draw count
// read off the targeted opponent's tally of damage taken this turn
// (every source, combat or not), read as the ability resolves. If the
// target has gone, the ability does nothing (CR 608.2b), and declining
// discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d834f1b-a4b5-4676-bb6d-2ffc6b3bccd2",
		Name:            "Knollspine Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.WhenThisEnters("Knollspine Dragon — discard your hand, then draw cards equal to the damage dealt to target opponent this turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					n := ctx.Game.DamageTakenThisTurn(ts[0].ID)
					if _, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller()); err != nil {
						return err
					}
					return effects.DrawCards{N: n}.Apply(ctx)
				}))),
				"Knollspine Dragon — discard your hand and draw cards equal to the damage dealt to the target opponent this turn?"),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
