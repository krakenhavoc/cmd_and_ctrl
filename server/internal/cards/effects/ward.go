package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward.go — CR 702.21, the keyword ADR 0038 §7 deliberately left out
// of the targeting choke point, built where it actually belongs.
//
// The ADR's argument is worth keeping in front of the code, because
// the tempting implementation is the wrong one. Ward LOOKS like
// hexproof: both appear on a creature, both make it awkward to
// remove, and hexproof is enforced in ten lines at
// game.CanBeTargetedBy. But CR 702.21a makes ward a TRIGGERED
// ability, and the difference is observable on every play:
//
//   - A warded permanent is a perfectly LEGAL target. The removal
//     spell is announced, it goes on the stack, and the trigger goes
//     on top of it. Enforcing ward at the targeting gate would
//     refuse the announce outright — no trigger, no payment offered,
//     no response window, and no way for the spell's controller to
//     pay and have it resolve.
//   - The payer is not the ward permanent's controller. It is
//     whoever cast the spell, which is exactly the shape
//     QueuePayUnlessForEffect already has: a chooser who is not the
//     controller of the source.
//   - The trigger is an object on the stack in its own right. It can
//     be responded to, and countering IT leaves the removal spell
//     alive and unpaid-for.
//
// PARAMETERISATION. The other half of ADR 0038 §7 is that ward
// carries a cost — "ward {2}", "ward—pay 3 life", "ward—sacrifice a
// creature" — and Characteristic.Abilities is a []string of bare
// tokens with nowhere to put it. That is why ward does NOT join
// canonicalKeywords here: a bare "ward" token in the enforced table
// would tell the coverage signal (ADR 0037) that every ward card is
// implemented while the engine had no idea what to charge. The cost
// lives where the rest of a card's parameters live — on the catalog
// Spec, as an argument to the helper below — and a ward card that is
// not in the catalog still flags as unimplemented, which is the
// honest answer.
//
// SCOPE. Only MANA wards ship. "Ward—Pay 3 life" (Sedgemoor Witch)
// and "Ward—Sacrifice a creature" (Vein Ripper) reuse the same
// trigger and the same consequence, but the payment is not a mana
// cost, and PendingChoicePayUnless parses its cost with ParseCost.
// A life-or-sacrifice ward needs a pay-unless prompt whose cost is
// an AbilityCost rather than a string; that is the next increment,
// and WardCost is where it lands.

// WardCost is the payment a ward demands. Today it is a mana cost
// string and nothing else; the type exists so that the day
// "ward—pay 3 life" ships, every card file already reads
// `Ward(WardMana("{2}"))` and only this file changes.
type WardCost struct {
	Mana string
}

// WardMana builds the "ward {N}" cost — the only printed form the
// engine can charge today.
func WardMana(cost string) WardCost { return WardCost{Mana: cost} }

// Ward builds the CR 702.21a triggered ability: "Whenever this
// permanent becomes the target of a spell or ability an opponent
// controls, counter that spell or ability unless its controller pays
// <cost>."
//
// Two details in the AppliesTo that are the rule rather than
// defensive coding:
//
//   - `ev.Actor != source.Controller` is "an OPPONENT controls".
//     Your own Giant Growth on your own warded creature does not
//     trigger it, which is the difference between ward and shroud.
//   - The trigger fires per target INSTANCE, because
//     EventBecomesTarget is emitted per target slot (CR 115.7). A
//     spell that targets the same warded creature twice triggers
//     ward twice, and the controller pays twice or the spell is
//     countered. That is correct and it falls out of the event
//     shape rather than needing anything here.
//
// The Effect reads the targeting spell off the event's StackItemID
// rather than its Source, because for an activated or triggered
// ability those differ — Source is the permanent the ability came
// from and would counter the wrong object (or nothing at all).
func Ward(cost WardCost, label string) game.TriggeredAbility {
	if label == "" {
		label = "Ward"
	}
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecomesTarget},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Actor != source.Controller
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			// Capture the two UUIDs, never the *Card or the *Game —
			// StackItem.Effect resolves against whichever Game it is
			// restored into.
			itemID, payer := ev.StackItemID, ev.Actor
			if itemID == uuid.Nil || payer == uuid.Nil {
				return nil
			}
			// The label is the card's own, unadorned: the cost is
			// already in it ("Hulking Raptor — ward {2}"), and the
			// pay-or-counter wording belongs on the PROMPT, which is
			// where the decision is actually made.
			return game.NewTriggeredItem(source, label,
				func(g *game.Game, item *game.StackItem) error {
					return wardPayOrCounter(g, item, itemID, payer, cost)
				})
		},
	}
}

// wardPayOrCounter is the trigger's resolution: offer the payment to
// the spell's controller, counter on a decline.
//
// The pay-unless prompt already handles the two ways a decline
// happens — answering "no", and answering "yes" without the mana in
// pool plus untapped sources — and treats them identically, which is
// CR 118.12.
//
// A spell that has already left the stack when the trigger resolves
// (countered in response, or its own controller responded by
// sacrificing the ward permanent and the spell then fizzled) makes
// the counter a no-op: CounterTargetForEffect returns
// ErrCardNotOnStack and there is nothing to charge for either, so
// the prompt is skipped entirely rather than asking a player to pay
// for a spell that no longer exists.
func wardPayOrCounter(g *game.Game, item *game.StackItem, targetingItem, payer uuid.UUID, cost WardCost) error {
	if g.StackItemForEffect(targetingItem) == nil {
		return nil
	}
	return g.QueuePayUnlessForEffect(payer, item.SourceCardID, cost.Mana,
		"Ward — pay "+cost.Mana+" or the spell is countered",
		func(g *game.Game) error {
			if g.StackItemForEffect(targetingItem) == nil {
				return nil
			}
			return g.CounterTargetForEffect(targetingItem)
		})
}
