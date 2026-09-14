package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// World Map — Artifact {1} (EDHREC rank 3070):
//
//	"{1}, {T}, Sacrifice this artifact: Search your library for a
//	 basic land card, reveal it, put it into your hand, then shuffle.
//	 {3}, {T}, Sacrifice this artifact: Search your library for a
//	 land card, reveal it, put it into your hand, then shuffle."
//
// Expedition Map with a cheaper basic-only mode. Two CR 602
// abilities on the same three cost components, each a search to
// hand with the reveal the printed text asks for; the two differ
// only in price and predicate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51d4f95e-aad2-4e5f-9f15-a68bbdf9ab3e",
		Name:         "World Map",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{1}, {T}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
				Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "World Map — a basic land card, to your hand",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "{3}, {T}, Sacrifice this artifact: Search your library for a land card, reveal it, put it into your hand, then shuffle.",
				Cost:  Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: b03IsLandCard,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "World Map — a land card, to your hand",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
