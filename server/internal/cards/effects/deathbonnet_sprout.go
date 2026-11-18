package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathbonnet Sprout // Deathbonnet Hulk — Creature — Fungus {G}, 1/1 //
// Creature — Fungus Horror 3/3 (Midnight Hunt, Innistrad: Double Feature):
//
//	Deathbonnet Sprout
//	  "At the beginning of your upkeep, mill a card. Then if there are three
//	   or more creature cards in your graveyard, transform this creature."
//	Deathbonnet Hulk
//	  "At the beginning of your upkeep, you may exile a card from a
//	   graveyard. If a creature card was exiled this way, put a +1/+1
//	   counter on this creature."
//
// The "then if" is not an intervening if: the trigger always happens, the
// mill goes first and the count is taken after it. The Hulk's exile is a
// pick over every graveyard, asked as the trigger resolves; the counter
// goes on only if the card exiled was a creature card, and a Hulk that has
// left the battlefield in response still lets you exile and gets nothing.
//
// No simplification.
func init() {
	const oracle = "c92d6f26-5856-42dc-9ee5-21292d78a1fb"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Deathbonnet Sprout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Deathbonnet Sprout — mill a card; then if you have three or more creature cards in your graveyard, transform it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (MillCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					if creatureCardsInGraveyard(g, item.Controller) < 3 {
						return nil
					}
					return TransformThis{}.Apply(ctx)
				}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Deathbonnet Hulk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Deathbonnet Hulk — you may exile a card from a graveyard; if it was a creature card, put a +1/+1 counter on this creature",
				deathbonnetHulkAsk),
		},
	})
}

// creatureCardsInGraveyard counts the creature cards in a player's graveyard.
func creatureCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			n++
		}
	}
	return n
}

// deathbonnetHulkAsk offers every card in every graveyard.
func deathbonnetHulkAsk(g *game.Game, item *game.StackItem) error {
	var cards []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Deathbonnet Hulk — exile a card from a graveyard (a creature card earns a +1/+1 counter), or none",
		Cards:    cards,
		Min:      0,
		Max:      1,
		// No Zone: the candidates span every seat's graveyard.
		ThenKey: deathbonnetHulkExile,
		Asked:   game.NewAskedItem(item),
	})
	return nil
}

// deathbonnetHulkExile exiles the pick and, if it was a creature card, puts
// the counter on. An on-disk key: never renamed, never reused.
var deathbonnetHulkExile = askedPick("choose-cards/deathbonnet-hulk-exile",
	func(ctx *Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			return ExileThenIfItWas{Target: id, Was: WasCreatureCard, Keyed: deathbonnetHulkCounter}.Apply(ctx)
		}
		return nil
	})

// deathbonnetHulkCounter is "if a creature card was exiled this way, put a
// +1/+1 counter on this creature". An on-disk key: never renamed, never reused.
var deathbonnetHulkCounter = IfItWasTail("if-it-was/deathbonnet-hulk-counter", func(ctx *Context) error {
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}, nil)
