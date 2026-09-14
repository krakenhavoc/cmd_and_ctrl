package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sweet Oblivion — Sorcery {1}{U}:
//
//	"Target player mills four cards.
//	 Escape—{3}{U}, Exile four other cards from your graveyard."
//
// The self-feeding half of the mechanic. Pointed at yourself it
// digs four cards deeper into the graveyard the escape cost is paid
// out of, which is the loop the whole Theros escape shell is built
// around — and pointed at an opponent it is ordinary mill.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c023538d-2feb-4af5-a44e-b355a190f081",
		Name:             "Sweet Oblivion",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{3}{U}", 4)},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return MillCards{Player: t.ID, N: 4}.Apply(ctx)
			}
			return nil
		},
	})
}
