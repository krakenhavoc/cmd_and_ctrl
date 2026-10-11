package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nether Traitor — Creature — Spirit {B}{B}, 1/1:
//
//		"Haste
//		 Shadow (This creature can block or be blocked by only creatures
//		 with shadow.)
//		 Whenever another creature is put into your graveyard from the
//		 battlefield, you may pay {B}. If you do, return this card from
//		 your graveyard to the battlefield."
//
//	  - Haste and shadow are printed keywords the engine already enforces.
//	  - The trigger watches from the GRAVEYARD only (InGraveyard, #925),
//	    the one zone where "return this card from your graveyard" can do
//	    anything. "You" is the card's owner (CR 108.4).
//	  - "Your graveyard" is the owner's: a creature goes to ITS owner's
//	    graveyard whoever controlled it, so a stolen creature that dies
//	    counts for its owner's Traitor and a creature of the Traitor's
//	    owner that died under an opponent's control still does. Any
//	    creature card or token counts, since a token is put into the
//	    graveyard before it ceases to exist.
//	  - The Traitor must already be in the graveyard when the other
//	    creature is put there. One that died in the same wipe, or after
//	    the creature, never saw it (netherTraitorWasThere).
//	  - "You may pay {B}" is asked when the ability RESOLVES (MayPay), not
//	    when it triggers (#2494), and a payment is only offered while the
//	    Traitor is still the graveyard object that triggered the ability.
//	    The card is checked again after the payment, so one that left in
//	    the meantime stays gone (CR 400.7).
//
// Sandbox simplification, declared: "at the same time" is read off the
// event batch (Event.Batch), and the batch is one whole resolution or
// state-based-action pass. A Traitor that dies EARLIER in the same
// spell or ability than the creature, as a separate instruction, is
// treated as dying together with it and does not see it; printed, it
// would already be in the graveyard. The wipes, combat damage and
// sacrifices the card is played against are all simultaneous and
// exact.
func init() {
	Register(Spec{
		OracleID:        "6f16c3ac-a9b8-47e6-b18b-ae37c74d44a0",
		Name:            "Nether Traitor",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"If it dies earlier in the same spell or ability as another creature, it still counts as dying together with it and doesn't see that creature."},
		PrintedKeywords: []string{"haste", "shadow"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventLTB, netherTraitorSeesACreatureArrive,
				"Nether Traitor — you may pay {B} to return it to the battlefield", netherTraitorMayPay)),
		},
	})
}

// netherTraitorSeesACreatureArrive is the trigger condition: another
// creature went from the battlefield to the Traitor's owner's
// graveyard, and the Traitor was in that graveyard first.
func netherTraitorSeesACreatureArrive(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	if !ok || dead.Owner != source.Owner {
		return false
	}
	return netherTraitorWasThere(g, source, ev)
}

// netherTraitorWasThere reports whether the Traitor sat in the
// graveyard before the creature arrived. Two events in one event batch
// happened at the same time as far as the rules go (CR 603.2c,
// Event.Batch), so a Traitor whose own trip to the graveyard is in the
// batch of the qualifying event died together with that creature and
// never saw it (CR 603.10a looks back only for what was on the
// battlefield). A Traitor that died in an earlier batch, this turn or
// before, was already there.
func netherTraitorWasThere(g *game.Game, source *game.Card, ev game.Event) bool {
	events := g.EventsThisTurn()
	for i := len(events) - 1; i >= 0; i-- {
		prev := events[i]
		if prev.Batch < ev.Batch {
			break
		}
		if prev.Kind == game.EventLTB && prev.CardID == source.InstanceID && prev.Batch == ev.Batch {
			return false
		}
	}
	return true
}

func netherTraitorMayPay(g *game.Game, item *game.StackItem) error {
	if !netherTraitorStillThere(g, item) {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{B}",
		Question: "Nether Traitor — pay {B} to return it to the battlefield?",
		OnPay: func(ctx *Context) error {
			if !netherTraitorStillThere(ctx.Game, ctx.Item) {
				return nil
			}
			return returnThisCardFromYourGraveyard(ctx.Game, ctx.Item)
		},
	}.Apply(NewContext(g, item))
}

// netherTraitorStillThere is "this card" in the graveyard: the very
// object that triggered, not a card of the same name that left and
// came back (CR 400.7).
func netherTraitorStillThere(g *game.Game, item *game.StackItem) bool {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	ref := item.SourceObject
	if ref.ID != item.SourceCardID || ref.ID == uuid.Nil {
		return true
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == ref.ID {
			return z.Cards[i].ObjectEpoch == ref.Epoch
		}
	}
	return false
}
