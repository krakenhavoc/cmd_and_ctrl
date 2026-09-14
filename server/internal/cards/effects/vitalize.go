package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vitalize — Instant {G} (EDHREC rank 3555):
//
//	"Untap all creatures you control."
//
// The one-mana mass untap. Every tapped creature the caster controls
// untaps; nobody else's does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd3226c6-bd47-4246-8ab3-01316c5d1809",
		Name:         "Vitalize",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b34UntapAllCreaturesYouControl(ctx.Game, item)
		},
	})
}
