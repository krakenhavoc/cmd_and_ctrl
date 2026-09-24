package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diffusion Sliver — Creature — Sliver {1}{U}, 1/1 (EDHREC rank
// 4273):
//
//	"Whenever a Sliver creature you control becomes the target of a
//	 spell or ability an opponent controls, counter that spell or
//	 ability unless its controller pays {2}."
//
// Ward {2} for a whole board, on a two-mana body. It is the reason a
// Sliver deck is hard to answer one card at a time: every removal
// spell costs two more, and two more on every spell adds up long
// before the board does.
//
// It is BUILT on the ward machinery, not merely similar to it. CR
// 702.21a defines ward as exactly this triggered ability, and the only
// difference here is the scope of the AppliesTo — "a Sliver creature
// you control" instead of "this permanent". Everything else is shared:
// the trigger goes on the stack ABOVE the targeting spell so the
// spell's controller gets a real pay-or-be-countered decision, the
// payment is charged to the SPELL's controller rather than to the
// Sliver's, and countering the trigger leaves the removal spell alive
// and unpaid for.
//
// Three details are the rule rather than defensive coding:
//
//   - "An OPPONENT controls" is `ev.Actor != source.Controller`. Your
//     own pump spell on your own Sliver does not tax you.
//   - It fires per target INSTANCE (CR 115.3), because
//     EventBecomesTarget is emitted per target slot. A spell that
//     targets two of your Slivers is taxed twice, and the controller
//     pays twice or it is countered.
//   - The Diffusion Sliver is itself a Sliver creature its controller
//     controls, so it protects itself. That is the printed text: there
//     is no "other".
//
// The targeting spell is read off the event's StackItemID rather than
// its Source, because for an activated or triggered ability those
// differ — Source is the permanent the ability came from and would
// counter the wrong object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "458ca9f4-4622-48e1-9ca1-6a8117188973",
		Name:         "Diffusion Sliver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41SliverYouControlBecameAnOpponentsTarget(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// Capture the two UUIDs, never the *Card or the *Game.
				itemID, payer := ev.StackItemID, ev.Actor
				if itemID == uuid.Nil || payer == uuid.Nil {
					return nil
				}
				return game.NewTriggeredItem(source, "Diffusion Sliver — counter it unless its controller pays {2}",
					func(g *game.Game, item *game.StackItem) error {
						return wardPayOrCounter(g, item, itemID, payer, WardMana("{2}"))
					})
			},
		}},
	})
}
