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
				// The targeting item is what the resolution reads off the
				// triggering event; an event naming none has nothing to
				// counter (the payer-nil check is inside the helper).
				if ev.StackItemID == uuid.Nil {
					return false
				}
				return b41SliverYouControlBecameAnOpponentsTarget(ev, source, g)
			},
			Key: "Diffusion Sliver — counter it unless its controller pays {2}",
			// ADR 0041 P9 (tier 4-2): declared on the row. The targeting
			// item and the payer are read at resolution off the item's
			// triggering event rather than captured when the ability
			// triggered, so a restored item charges the same player for
			// the same spell.
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				ev := item.Trigger.Event
				return wardPayOrCounter(g, item, ev.StackItemID, ev.Actor, WardMana("{2}"))
			},
		}},
	})
}
