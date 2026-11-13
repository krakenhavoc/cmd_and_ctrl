package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Time of Need — Sorcery {1}{G} (Champions of Kamigawa, #5603):
//
//	"Search your library for a legendary creature card, reveal it, put it
//	 into your hand, then shuffle."
//
// Captain Sisay's search narrowed to creature cards: "legendary" and
// "creature" are read off the card in the library (CR 701.23a), and the
// card found is revealed to the table (CR 701.20a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f2adfd0-c8ab-4bcc-ae55-cb0e798aec7f",
		Name:         "Time of Need",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsLegendary() && c.IsCreature() },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Time of Need — choose a legendary creature card to put into your hand",
			}.Apply(ctx)
		},
	})
}
