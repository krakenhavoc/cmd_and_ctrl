package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marketback Walker — Artifact Creature — Construct {X}{X}, 0/0
// (EDHREC rank 4059):
//
//	"This creature enters with X +1/+1 counters on it.
//	 {4}: Put a +1/+1 counter on this creature.
//	 When this creature dies, draw a card for each +1/+1 counter on
//	 it."
//
// A colourless mana sink that turns into cards. {X}{X} is a terrible
// rate on the way in — six mana for a 3/3 — and that is not what the
// card is for: it is for a late game with nothing to do, where the
// Walker grows a counter per four mana and every counter is a card
// the moment somebody kills it. In a deck with a sacrifice outlet it
// is Ambition's Cost you can cash in whenever you like.
//
// {X}{X} IS TWO X SLOTS, so casting it for X=3 costs six mana and
// puts three counters on it. The engine charges both slots and the
// entry clause reads the single announced X, which is the printed
// arithmetic.
//
// THE DEATH TRIGGER READS THE LAST-KNOWN COUNTERS, not the card in
// the graveyard, which has none (CR 400.7 clears them the moment it
// moves). CR 603.10 has a dies-trigger look at the game state just
// before the permanent left, and b13LastKnownCounters is that read,
// reconstructed from the event log. So a Walker that entered with
// three counters, grew two, and then died draws five — and a Walker
// that had its counters removed in response draws only what it still
// held.
//
// The activated ability is a plain {4} with no tap, so it can be used
// at instant speed, any number of times, and the turn the Walker
// lands. That is the printed text and it is what makes the card a
// mana sink rather than a creature.
//
// The X counters are the printed CR 614.1c entry clause and ride the
// CR 614 pipeline as one — XCounters, seeded onto the entry event off
// the resolving stack item while the spell is still there (#1002).
// They land on the PERMANENT, after the move and before EventETB, so
// Doubling Season and Hardened Scales apply, the card's own enters
// trigger reads a finished creature, and a "whenever one or more
// counters are put on a permanent you control" payoff sees them —
// which it could not while they went onto a card still on the stack.
//
// Cast for X=0 the Walker enters as the printed 0/0 it is and the
// next state-based check puts it into its owner's graveyard (CR
// 704.5f) before {4} can grow it, as in paper. CR 601.2b allows the
// announcement; it simply does not survive it. It used to survive
// here and was declared to players as the one way the card played
// STRONGER than printed; #691 took both the gap and the caveat away.
func init() {
	Register(Spec{
		OracleID:                   "0405e0a9-6d02-4691-bdb8-59c72b824dab",
		Name:                       "Marketback Walker",
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Completeness:               CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{4}: Put a +1/+1 counter on this creature.",
			Cost:   ManaCost("{4}"),
			Effect: putCounterOnSelf,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Key: "Marketback Walker — draw a card for each +1/+1 counter on it",
			// The dies trigger reads the LAST-KNOWN counter count, which
			// is a fact about the moment it died (ADR 0041 P9's fill-in
			// Build), not something the resolving item can re-derive
			// from a board the permanent has already left.
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Marketback Walker — draw a card for each +1/+1 counter on it")
				item.Params.Amount = b13LastKnownCounters(g, source.InstanceID, game.CounterPlusOne)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return DrawCards{Player: item.Controller, N: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
