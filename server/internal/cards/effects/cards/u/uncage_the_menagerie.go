package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uncage the Menagerie — Sorcery {X}{G}{G} (Hour of Devastation, #5534):
//
//	"Search your library for up to X creature cards with different names that
//	 each have mana value X, reveal them, put them into your hand, then
//	 shuffle."
//
// Ecological Appreciation's search with its different-names set rule (one
// card of each name) and a count of X rather than four: the creatures have
// mana value exactly X, read as the card each is in the library, and at most
// X are taken. With X = 0 the search takes nothing but the library is still
// shuffled. The cards are revealed and put into your hand together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2043685-f091-444f-9c35-d238afa24d7b",
		Name:         "Uncage the Menagerie",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
			}
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() == x },
				Dest:      game.ZoneHand,
				Limit:     x,
				UpTo:      true,
				Reveal:    true,
				Shuffle:   true,
				Rule:      effects.SearchRuleDifferentNames,
				Reason:    "Uncage the Menagerie — up to X creature cards with different names and mana value X",
			}.Apply(ctx)
		},
	})
}
