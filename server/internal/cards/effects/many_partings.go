package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Many Partings — Sorcery {G} (EDHREC rank 1957):
//
//	"Search your library for a basic land card, reveal it, put it
//	 into your hand, then shuffle. Create a Food token."
//
// A one-mana Lay of the Land that leaves a Food behind. The search
// is the searcher's prompt when the library holds more than one
// basic; the Food is the shared artifact-token template. The Food
// is created in the search's continuation so it lands after the
// land is in hand, in the printed order, whether or not a prompt
// was needed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92ea750e-62b7-4422-a527-ddf435759d08",
		Name:         "Many Partings",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Many Partings — a basic land card, revealed, to hand",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
				},
			}.Apply(ctx)
		},
	})
}
