package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nomads' Assembly — Sorcery {4}{W}{W}:
//
//	"Create a 1/1 white Kor Soldier creature token for each creature
//	 you control.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The creatures are counted once, as the spell resolves, and the
// tokens are made together. Rebound is the engine's keyword
// (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9a0088a-3c86-4baf-9754-79006f733ce1",
		Name:            "Nomads' Assembly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ctx.Game.RecomputeLayersIfStaleLocked()
			n := len(permanentsControlledByMatching(ctx.Game, ctx.Controller(), Creature()))
			if n == 0 {
				return nil
			}
			return CreateToken{Template: TokenCard("1/1 white Kor Soldier"), N: n}.Apply(ctx)
		},
	})
}
