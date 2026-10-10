package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Germinate Recruits — Instant {2}{W}:
//
//	"Create X 2/2 colorless Wizard Soldier creature tokens named
//	 Cadet, where X is the amount of life you gained this turn."
//
// X is the controller's life-gained tally for the turn, read as the
// spell resolves (it counts life gained before the spell was cast and
// any gained in response). The Cadet row is Reality Fracture's named
// 2/2 Wizard Soldier token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c3a1313-a66e-4317-9732-e4ff9288b0db",
		Name:         "Germinate Recruits",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := b15LifeGainedThisTurn(ctx.Game, ctx.Controller())
			if x <= 0 {
				return nil
			}
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("2/2 colorless Wizard Soldier named Cadet"),
				N:          x,
			}.Apply(ctx)
		},
	})
}
