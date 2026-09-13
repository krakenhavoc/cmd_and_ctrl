package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampant Growth — Sorcery {1}{G}:
//
//	"Search your library for a basic land card, put it onto the
//	battlefield tapped, then shuffle."
//
// Cultivate's smaller cousin: one search, tapped, shuffle. The
// controller picks which basic (S22) — a library holding exactly one
// basic skips the prompt, because a modal with one button is worse
// than no modal.
func init() {
	Register(Spec{
		OracleID:     "8539f295-5d58-4436-a73a-b9277c4c7795",
		Name:         "Rampant Growth",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Rampant Growth — a basic land",
			}.Apply(ctx)
		},
	})
}
