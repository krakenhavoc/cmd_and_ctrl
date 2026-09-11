package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expedition Map — Artifact {1} (EDHREC rank 457):
//
//	"{2}, {T}, Sacrifice this artifact: Search your library for a
//	 land card, reveal it, put it into your hand, then shuffle."
//
// The one-mana land tutor — Cabal Coffers, Urborg, Field of the
// Dead, Gaea's Cradle, whichever land the deck is built around.
// Three cost components (Myriad Landscape's), a search to HAND with
// the reveal the printed text asks for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "8fcf50cd-e6d0-4516-850f-d42ee75dcc3a",
		Name:     "Expedition Map",
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this artifact: Search your library for a land card, reveal it, put it into your hand, then shuffle.",
			Cost:  Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b03IsLandCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Expedition Map — a land card, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
