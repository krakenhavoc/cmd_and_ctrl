package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hour of Reckoning — Sorcery {4}{W}{W}{W} (EDHREC rank 1032):
//
//	"Convoke
//	 Destroy all nontoken creatures."
//
// The token deck's one-sided wrath: the tokens help pay for it and
// then survive it. Convoke is the S22 tap cost (Chord of Calling's),
// and the sweep is DestroyAllMatching over "creature that isn't a
// token" — one simultaneous event, so a Blood Artist sees every death
// at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3bc13640-03f8-4b19-b0a4-7e7cb5271c0c",
		Name:         "Hour of Reckoning",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), Not(IsTokenPredicate()))}.Apply(ctx)
		},
	})
}
