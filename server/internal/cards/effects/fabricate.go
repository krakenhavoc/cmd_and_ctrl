package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fabricate — Sorcery {2}{U} (EDHREC rank 470):
//
//	"Search your library for an artifact card, reveal it, put it
//	 into your hand, then shuffle."
//
// Blue's artifact tutor: Demonic Tutor's shape with the predicate
// narrowed to artifact cards and the reveal the printed text asks
// for. The S22 search chooser lets the caster pick which artifact
// (or fail to find).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "422e1869-134f-463d-9fa1-86b66a998b3e",
		Name:         "Fabricate",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: b03IsArtifactCard,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Fabricate — an artifact card, to your hand",
			}.Apply(ctx)
		},
	})
}
