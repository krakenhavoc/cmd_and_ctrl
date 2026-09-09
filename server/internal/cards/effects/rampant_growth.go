package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampant Growth — Sorcery {1}{G}:
//
//	"Search your library for a basic land card, put it onto the
//	battlefield tapped, then shuffle."
//
// Cultivate's smaller cousin: one search, tapped, shuffle. Sandbox
// simplification inherited from SearchLibrary — the first basic in
// library order is taken rather than the controller choosing, so a
// deck with mixed basics gets an arbitrary (but deterministic) one.
// The pick UI is the S22 deferral that covers every tutor.
func init() {
	Register(Spec{
		OracleID: "8539f295-5d58-4436-a73a-b9277c4c7795",
		Name:     "Rampant Growth",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
			}.Apply(ctx)
		},
	})
}
