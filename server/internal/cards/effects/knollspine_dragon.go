package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "9d834f1b-a4b5-4676-bb6d-2ffc6b3bccd2",
		Name:            "Knollspine Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(Optional(WhenThisEnters("Knollspine Dragon — discard your hand, then draw cards equal to the damage dealt to target opponent this turn",
				Do(Instruction(func(ctx *Context) error {
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					n := ctx.Game.DamageTakenThisTurn(ts[0].ID)
					if _, err := discardWholeHand(ctx.Game, ctx.Controller()); err != nil {
						return err
					}
					return DrawCards{N: n}.Apply(ctx)
				}))),
				"Knollspine Dragon — discard your hand and draw cards equal to the damage dealt to the target opponent this turn?"),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
