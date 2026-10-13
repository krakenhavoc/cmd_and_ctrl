package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kodama's Reach — Sorcery — Arcane {2}{G}:
//
//	"Search your library for up to two basic land cards, reveal those
//	cards, put one onto the battlefield tapped and the other into
//	your hand, then shuffle."
//
// Functionally identical to Cultivate (the Arcane subtype matters
// only to splice, which the engine has no notion of), so it shares
// cultivateSearch: one optional search per destination, the first
// deferring its shuffle, the second shuffling once. "Up to two" is
// honoured (#2942) — see cultivate.go.
func init() {
	Register(Spec{
		OracleID:     "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90",
		Name:         "Kodama's Reach",
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return cultivateSearch(ctx, "Kodama's Reach")
		},
	})
}
