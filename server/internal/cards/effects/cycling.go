package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cycling.go — #660: cycling and typecycling (CR 702.29).
//
// "Cycling {2}" is not an alternative way to cast the card, and the
// #94 body and five card files in this package said it was. It is an
// ACTIVATED ABILITY that functions only while the card is in its
// owner's hand:
//
//	"{2}, Discard this card: Draw a card."   CR 702.29a
//
// Typecycling (CR 702.29e) is the same ability with a different
// effect: "Basic landcycling {2}" is "{2}, Discard this card: Search
// your library for a basic land card, reveal it, put it into your
// hand, then shuffle."
//
// Both come out of one constructor pair, and both are ordinary CR 602
// activations — same stack item, same Effect closure, same cost
// validation — reaching the hand through ActivatedAbilityShape.Zones
// (ADR 0062 Decision 1) and paying through AbilityCost.DiscardSelf
// (Decision 2). Nothing in the engine knows the word "cycling" except
// the one declarative bit that makes EventCycle fire.

// Cycling is "Cycling <cost>" — CR 702.29a. `cost` is the printed
// mana cost of the ability, "{2}" or "{1}{R}"; the discard and the
// hand-only restriction come with the keyword and are not the card
// file's to spell out.
//
// Give the ability to the card the way its oracle text reads:
//
//	Activated: []ActivatedAbility{Cycling("{3}")},
//
// A card with cycling and nothing else is otherwise a plain Spec —
// the whole point of the constructor is that a Triome's cycling costs
// it one line.
func Cycling(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:   "Cycling " + cost + " (" + cost + ", Discard this card: Draw a card.)",
		Cost:    Plus(ManaCost(cost), DiscardThis()),
		Zones:   []game.ZoneKind{game.ZoneHand},
		Cycling: true,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
		},
	}
}

// Typecycling is "<Type>cycling <cost>" — CR 702.29e. `keyword` is
// the printed word ("Basic landcycling", "Plainscycling",
// "Slivercycling"), `what` is the clause as the card prints it ("a
// basic land card"), and `pred` decides which library cards match.
//
// The ability is a cycling ability (CR 702.29e says so explicitly), so
// it emits EventCycle and a "whenever you cycle a card" watcher sees
// it — which is why this shares Cycling's body rather than being a
// second kind of thing that happens to discard.
//
// The search is "search … reveal … put into your hand, then shuffle",
// which is CR 702.29e word for word: an unsuccessful search still
// shuffles, and the reveal is what stops a typecycling from being a
// hidden tutor.
func Typecycling(keyword, cost, what string, pred func(game.Card) bool) ActivatedAbility {
	ab := Cycling(cost)
	ab.Label = keyword + " " + cost + " (" + cost + ", Discard this card: Search your library for " +
		what + ", reveal it, put it into your hand, then shuffle.)"
	reason := keyword + " — " + what + ", to your hand"
	ab.Effect = func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: pred,
			Dest:      game.ZoneHand,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
			Reason:    reason,
		}.Apply(NewContext(g, item))
	}
	return ab
}

// BasicLandcycling is Typecycling's commonest printing — Sylvan
// Reclamation, the Zendikar cycling lands, the Ikoria cycle.
func BasicLandcycling(cost string) ActivatedAbility {
	return Typecycling("Basic landcycling", cost, "a basic land card", IsBasicLand)
}

// WheneverYouCycle is "Whenever you cycle a card, …" — Astral Slide,
// Drake Haven, Fluctuator's cousins. The source watches EventCycle
// and fires when the CONTROLLER of the watching permanent is the
// player who cycled (CR 702.29b: the cycling player is the one who
// activated the ability).
//
// It works from the battlefield today with no change to the
// harvester, because the battlefield scan is the harvester's first
// pass. "When you cycle THIS card" (Magmakin Artillerist) is a
// different trigger and does NOT work yet: the card is in the
// graveyard by the time the event fires (CR 702.29c) and the
// harvester has no scan that finds it there. ADR 0062 Decision 7
// carries the design and the reason it waits.
func WheneverYouCycle(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCycle, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source != nil && ev.Actor == source.Controller
	}, label, effect)
}
