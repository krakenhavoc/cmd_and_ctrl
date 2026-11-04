package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donatello's Technique — Sorcery {2}{U} (TMT, #4457):
//
//	"Sneak {U} (You may cast this spell for {U} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Draw two cards."
//
// Sneak lets a sorcery be cast in its caster's declare blockers step
// (CR 702.190a); the body is Divination's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "194284f7-cf92-4d9f-a8ce-e4dc3505ab7a",
		Name:             "Donatello's Technique",
		Purpose:          game.Purpose{Draws: 2},
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
