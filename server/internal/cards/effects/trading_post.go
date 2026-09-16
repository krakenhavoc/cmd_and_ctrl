package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trading Post — Artifact {4} (EDHREC rank 2140):
//
//	"{1}, {T}, Discard a card: You gain 4 life.
//	 {1}, {T}, Pay 1 life: Create a 0/1 white Goat creature token.
//	 {1}, {T}, Sacrifice a creature: Return target artifact card from
//	 your graveyard to your hand.
//	 {1}, {T}, Sacrifice an artifact: Draw a card."
//
// The artifact deck's Swiss-army knife. Three of the four abilities
// are CR 602 activated abilities built from the cost components the
// engine has — mana, tap, life, sacrifice-a-creature (the Goat it
// just made qualifies, as printed) and sacrifice-an-artifact (the
// Post itself qualifies, as printed). The regrowth targets an
// artifact card the controller owns, picked in the zone browser and
// re-checked at resolution.
//
// Sandbox simplification, declared: the first ability, "{1}, {T},
// Discard a card: You gain 4 life", is NOT registered. An activated
// ability's cost has no discard component (AbilityCost carries tap,
// sacrifice-this, sacrifice-another, mana, life and loyalty — the
// discard cost the engine has is the cast-time AdditionalCost, a
// different pipeline), and shipping the ability with the discard
// omitted would be stronger than printed. Three abilities of four,
// weaker and never stronger; the lifegain lands with the discard
// cost component.
func init() {
	postCost := Plus(ManaCost("{1}"), TapCost())
	Register(Spec{
		OracleID:     "63788566-e25a-44bb-bb55-197e1b93b3e8",
		Name:         "Trading Post",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"{1}, {T}, Discard a card: You gain 4 life\" ability isn't implemented — the other three abilities work."},
		Activated: []ActivatedAbility{
			{
				Label: "{1}, {T}, Pay 1 life: Create a 0/1 white Goat",
				Cost:  Plus(postCost, PayLife(1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("0/1 white Goat"), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{1}, {T}, Sacrifice a creature: Return target artifact card from your graveyard to your hand",
				Cost:    Plus(postCost, SacrificeACreature()),
				Targets: TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
				},
			},
			{
				Label: "{1}, {T}, Sacrifice an artifact: Draw a card",
				Cost:  Plus(postCost, b10SacrificeAnArtifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
