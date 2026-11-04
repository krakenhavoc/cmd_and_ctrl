package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Courier of Comestibles — Creature — Human Citizen {1}{G}, 1/2
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"When this creature enters, you may search your library for a Food
//	 card, reveal it, put it into your hand, then shuffle. If you don't
//	 put a card into your hand this way, create a Food token."
//
// A "you may search" (SearchLibrary.Optional) whose rest is a keyed
// tail told the cards found. The tail runs on every answer: a declined
// search (nothing happened, no shuffle, #4134), a search that found
// nothing (it is a stated quality, so it may, CR 701.23b), and a found
// Food card. Only the last puts a card into the hand, so the first two
// create the Food token. "A Food card" is a card with the Food subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "049d8b06-db90-4b9c-8195-d987b3ef1005",
		Name:         "Courier of Comestibles",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Courier of Comestibles — you may search for a Food card; if you don't, create a Food token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.HasSubtype("Food") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Courier of Comestibles — a Food card",
						Source:    item.SourceCardID,
						ThenTail:  ctx.ItemTail(courierOfComestiblesFood),
					}.Apply(ctx)
				}),
		},
	})
}

// courierOfComestiblesFood is "if you don't put a card into your hand
// this way, create a Food token". An on-disk key.
var courierOfComestiblesFood = CardTail("courier-of-comestibles/food-unless-found", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) > 0 {
		return nil
	}
	return CreateToken{Template: FoodToken(), N: 1}.Apply(ctx)
})
