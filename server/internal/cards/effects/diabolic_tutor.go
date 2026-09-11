package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diabolic Tutor — Sorcery {2}{B}{B} (EDHREC rank 482):
//
//	"Search your library for a card, put that card into your hand,
//	 then shuffle."
//
// Demonic Tutor at twice the price — the same unconditional search
// with a nil predicate, so every card in the library is a candidate
// and the controller picks one through the S22 search chooser.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "14589b6b-1814-46f9-a364-83cc15dacac2",
		Name:     "Diabolic Tutor",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Diabolic Tutor — search your library for a card",
			}.Apply(ctx)
		},
	})
}
