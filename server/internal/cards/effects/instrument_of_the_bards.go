package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Instrument of the Bards — Legendary Artifact {G} (Adventures in the
// Forgotten Realms):
//
//	"At the beginning of your upkeep, you may put a harmony counter on
//	 Instrument of the Bards.
//	 {3}{G}, {T}: Search your library for a creature card with mana value
//	 equal to the number of harmony counters on Instrument of the Bards,
//	 reveal it, and put it into your hand. If that card is legendary,
//	 create a Treasure token. Then shuffle."
//
// Aether Vial's counter and its read of the counters as the ability
// resolves (last-known information if the Instrument has left, CR
// 608.2h). A creature card with {X} in its cost has mana value 0 in the
// library (CR 202.3e). The Treasure is made only when a card was found
// and it is legendary; the shuffle happens either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d79755c4-3c61-4fc8-8eb7-fbdc23951748",
		Name:         "Instrument of the Bards",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Instrument of the Bards — put a harmony counter on it",
				Do(CounterOnThis{Kind: "harmony", N: 1})),
				"Instrument of the Bards — put a harmony counter on it?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, {T}: Search your library for a creature card with mana value equal to the number of harmony counters on this artifact, reveal it, and put it into your hand. If that card is legendary, create a Treasure token. Then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{G}"), TapCost()),
			Effect:  instrumentOfTheBardsSearch,
		}},
	})
}

func instrumentOfTheBardsSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 0
	if inst, ok := ctx.SourcePermanent(); ok {
		n = inst.Counters["harmony"]
	}
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return c.IsCreature() && b03ManaValueIs(n)(g, item.Controller, c) },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Instrument of the Bards — a creature card with mana value equal to its harmony counters",
		ThenTail:  ctx.Tail(instrumentOfTheBardsTreasure),
	}.Apply(ctx)
}

// instrumentOfTheBardsTreasure makes the Treasure when the found card is
// legendary. An on-disk key.
var instrumentOfTheBardsTreasure = CardTail("instrument-of-the-bards/treasure-if-legendary", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !c.IsLegendary() {
		return nil
	}
	return CreateToken{Controller: o.Tail.Player, Template: TreasureToken(), N: 1}.Apply(ctx)
})
