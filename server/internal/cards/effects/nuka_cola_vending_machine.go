package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nuka-Cola Vending Machine — Artifact {3} (EDHREC rank 1294):
//
//	"{1}, {T}: Create a Food token. (It's an artifact with "{2}, {T},
//	 Sacrifice this token: You gain 3 life.")
//	 Whenever you sacrifice a Food, create a tapped Treasure token.
//	 (It's an artifact with "{T}, Sacrifice this token: Add one mana
//	 of any color.")"
//
// The Food deck's Treasure engine. The activation is a CR 602
// ability with a mana-and-tap cost; the trigger watches
// EventSacrifice, which fires before the zone move so the Food is
// still on the battlefield to be recognised by subtype (a Gingerbrute
// or a Food-typed token from another card counts, as printed). The
// Treasure enters tapped through CreateTokenAdvanced, and the Amulet
// of Vigor in the same batch sees that entry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cfb03e5-aca6-4fe2-a3f1-93e1f0cbf9e1",
		Name:         "Nuka-Cola Vending Machine",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Create a Food token.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventSacrifice},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11SacrificedAFood(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Nuka-Cola Vending Machine — create a tapped Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateTokenAdvanced{
							Controller: item.Controller,
							Spec:       Token(TreasureToken()).EntersTapped(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
