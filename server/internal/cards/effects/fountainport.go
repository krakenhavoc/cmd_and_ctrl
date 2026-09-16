package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fountainport — Land (EDHREC rank 928):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice a token: Draw a card.
//	 {3}, {T}, Pay 1 life: Create a 1/1 blue Fish creature token.
//	 {4}, {T}: Create a Treasure token."
//
// The Bloomburrow utility land: a colorless source that turns spare
// mana into a token, a Treasure, or — with a token to feed it — a
// card. Three activated abilities on one land, each with its own
// cost shape, all through the ordinary CR 602 path: mana and life
// validated before anything is paid, the token sacrifice picked from
// the activator's own tokens (any token — a Treasure, a Fish, a
// creature), and the tap shared, so the land does one thing a turn.
// The Fish is a real 1/1 blue creature token and the Treasure is
// tokens.go's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "94e8b0a9-44a1-4dce-8d44-78681ae638a1",
		Name:          "Fountainport",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{
			{
				Label: "{2}, {T}, Sacrifice a token: Draw a card.",
				Cost:  Plus(ManaCost("{2}"), TapCost(), b08SacrificeAToken()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "{3}, {T}, Pay 1 life: Create a 1/1 blue Fish creature token.",
				Cost:  Plus(ManaCost("{3}"), TapCost(), PayLife(1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 blue Fish"), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "{4}, {T}: Create a Treasure token.",
				Cost:  Plus(ManaCost("{4}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
