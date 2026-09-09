package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mary Read and Anne Bonny — 3/3 Legendary Creature — Human
// Assassin Pirate for {1}{U}{R}:
//
//	"Haste
//	 {T}: Draw a card, then discard a card.
//	 Whenever you discard an Island, Pirate, or Vehicle card, create
//	 a tapped Treasure token."
//
// The deck's commander and the engine that drives it: the tap
// ability loots every turn, and discarding the right card type turns
// that loot into mana. Both halves are expressible today —
// the activated ability through S21's cost model, the trigger
// through EventDiscardCard, which carries the discarded card so the
// type check reads it out of the graveyard.
//
// The Treasure enters tapped, so it's mana for the NEXT turn rather
// than a free ritual on the turn you loot.
func init() {
	Register(Spec{
		OracleID:        "5182de2d-aceb-450e-bd20-8bc7db124334",
		Name:            "Mary Read and Anne Bonny",
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label: "{T}: Draw a card, then discard a card",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return discardedByYou(ev, source) &&
					discardedCardHasType(ev, g, "island", "pirate", "vehicle")
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Mary Read and Anne Bonny — create a tapped Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TappedTreasureToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
