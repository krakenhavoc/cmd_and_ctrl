package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sprout — Instant {G} (Time Spiral, #5596):
//
//	"Create a 1/1 green Saproling creature token."
//
// One 1/1 green Saproling token for its caster, at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "424c6f5e-b386-47e9-b3fe-25b263097d40",
		Name:         "Sprout",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
		},
	})
}
