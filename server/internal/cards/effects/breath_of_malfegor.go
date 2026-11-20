package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breath of Malfegor — Instant {3}{B}{R} (Alara Reborn, #5586):
//
//	"Breath of Malfegor deals 5 damage to each opponent."
//
// Boltwave's body at 5: one damage instance to every opponent (CR 615.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81e6ade9-a38f-4655-b09c-9ca41a11face",
		Name:         "Breath of Malfegor",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageToEachOpponent(ctx.Game, item, 5)
		},
	})
}
