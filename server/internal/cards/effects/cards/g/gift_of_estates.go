package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gift of Estates — Sorcery {1}{W} (Strixhaven Mystical Archive, #5503):
//
//	"If an opponent controls more lands than you, search your library
//	 for up to three Plains cards, reveal them, put them into your hand,
//	 then shuffle."
//
// The condition is read as the spell resolves: some one opponent
// controls strictly more lands than you do. The search is for cards with
// the Plains subtype, not only basic ones (Keeper of the Accord's search
// is basic-only), up to three, revealed, to hand, then a shuffle. When
// no opponent is ahead, nothing happens and nothing is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bd70be6-752c-4ecb-a3e8-2bacec4b94fc",
		Name:         "Gift of Estates",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			you := ctx.Controller()
			ahead := false
			for _, opp := range g.Seats {
				if opp != nil && opp.ID != you && effects.B11OpponentControlsMoreThanYou(g, you, opp.ID, game.Card.IsLand) {
					ahead = true
				}
			}
			if !ahead {
				return nil
			}
			return effects.SearchLibrary{
				Player: you,
				Predicate: func(c game.Card) bool {
					return c.HasSubtypeWord("Plains")
				},
				Dest:    game.ZoneHand,
				Limit:   3,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Gift of Estates — up to three Plains cards",
			}.Apply(effects.NewContext(g, item))
		},
	})
}
