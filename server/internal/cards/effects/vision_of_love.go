package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vision of Love — Instant {1}{R}:
//
//	"You may sacrifice an artifact or discard a card. If you do, draw two
//	 cards."
//
// #4566's either/or may-pay as a spell resolves: one prompt naming the
// branch, a branch the caster can't pay is not offered, and the two cards
// are drawn only once the chosen branch is paid (CR 118.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70c96b50-ac2b-493d-bd15-fc66eb767111",
		Name:         "Vision of Love",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return mayPayArtifactOrCard(ctx, "Vision of Love — sacrifice an artifact or discard a card to draw two cards?", visionOfLoveDraw)
		},
	})
}

// visionOfLoveDraw is "If you do, draw two cards".
var visionOfLoveDraw = PayUnlessBranches("pay-unless/vision-of-love-draw-two",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
	}, nil)
