package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Thief — Creature — Human Rogue {2}{U}, 2/3 (Journey into Nyx, S82 OLD,
// #5558):
//
//	"Inspired — Whenever this creature becomes untapped, you may exchange
//	 control of target nonland permanent you control and target permanent an
//	 opponent controls that shares a card type with it."
//
// Puca's Mischief's "you may" exchange over two targets, with the second tied
// to the first by a shared card type (SharesACardTypeWithTarget), judged as the
// trigger goes on the stack and again as it resolves. The exchange is all or
// nothing (CR 701.12a). Inspired is Deepwater Hypnotist's untap watch on the
// creature itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e01289ea-634b-4be5-8b8e-5e1179761a9b",
		Name:         "Daring Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventUntapCard, effects.Self,
					"Daring Thief — you may exchange control of target nonland permanent you control and target permanent an opponent controls that shares a card type with it (inspired)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						mine, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						theirs, ok := ctx.ClauseTarget(1)
						if !ok {
							return nil
						}
						return effects.ExchangeControl{A: mine.ID, B: theirs.ID, Label: "Daring Thief — exchange control"}.Apply(ctx)
					}),
				effects.Clauses(
					effects.TargetPermanent("target nonland permanent you control", effects.Nonland(), effects.YouControl()),
					effects.SharesACardTypeWithTarget(
						effects.TargetPermanent("target permanent an opponent controls that shares a card type with it", effects.OpponentControls()), 0),
				)),
				"Daring Thief — exchange control?"),
		},
	})
}
