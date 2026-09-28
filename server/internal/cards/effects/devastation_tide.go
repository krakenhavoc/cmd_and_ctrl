package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devastation Tide — Sorcery {3}{U}{U} (#1665):
//
//	"Return all nonland permanents to their owners' hands.
//	 Miracle {1}{U} (You may cast this card for its miracle cost when
//	 you draw it if it's the first card you drew this turn.)"
//
// Evacuation widened to every nonland permanent, at sorcery speed —
// except on the first draw of a turn, when the miracle cast ignores
// timing and costs two. Symmetric: your own permanents come back too,
// and tokens cease to exist.
func init() {
	Register(Spec{
		OracleID:         "4245ee98-2d4c-49d1-8d07-80760cae2bf9",
		Name:             "Devastation Tide",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Miracle("{1}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{Match: Nonland()}.Apply(ctx)
		},
	})
}
