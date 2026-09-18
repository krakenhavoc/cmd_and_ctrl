package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Looter — 1/1 Creature — Merfolk Rogue for {1}{U} (EDHREC
// rank 4375):
//
//	"{T}: Draw a card, then discard a card."
//
// The card the verb is named after. Two mana for a repeatable filter
// every turn is the engine underneath every reanimator and every
// madness deck, and it has never been reprinted out of relevance.
// Roadmap batch 42 (#449), "no new machinery".
//
// Draw THEN discard — a loot, not a rummage — and the order is the
// whole card: you see the new card before you decide what to pitch.
// lootOne draws first and queues the discard prompt second, which is
// why the discard has no Then clause hanging off it; there is nothing
// printed after it.
//
// The discard is the controller's own choice (CR 701.8a), not the
// "at random" primitive.
//
// A tap cost, so the Looter cannot be used the turn it arrives
// (CR 302.6 summoning sickness) and untaps to do it again every turn
// after — both of which the engine's cost validation handles, not
// this file.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67362406-b1ca-49e2-800d-9050bfe8742a",
		Name:         "Merfolk Looter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{T}: Draw a card, then discard a card.",
			Cost:   TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
		}},
	})
}
