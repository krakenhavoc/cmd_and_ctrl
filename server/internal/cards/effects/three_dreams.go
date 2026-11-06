package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Three Dreams — Sorcery {4}{W} (Avatar: The Last Airbender Eternal, #5075):
//
//	"Search your library for up to three Aura cards with different names,
//	 reveal them, put them into your hand, then shuffle."
//
// One search with a stated quality (CR 701.23b), so it may come out short
// or find nothing, and the picked set must have different names: the
// engine's registered different-names set rule (search/different-names),
// Fblthp's and Emergent Ultimatum's. "Aura card" is the printed type line,
// as every library search reads it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3f9da94-c822-49b8-a1fc-be82189206af",
		Name:         "Three Dreams",
		Purpose:      game.Purpose{Tutors: 3},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Aura") },
				Dest:      game.ZoneHand,
				Limit:     3,
				Reveal:    true,
				Shuffle:   true,
				Rule:      searchRuleDifferentNames,
				Reason:    "Three Dreams — search for up to three Aura cards with different names",
			}.Apply(ctx)
		},
	})
}
