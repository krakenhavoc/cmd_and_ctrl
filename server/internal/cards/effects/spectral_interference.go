package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectral Interference — Instant {1}{U} (Aetherdrift, #4464):
//
//	"Counter target artifact or creature spell unless its controller
//	 pays {4}."
//
// Spell Pierce's pay-unless (CounterUnlessPaid) with Hurl into
// History's "artifact or creature spell" target. The prompt goes to the
// spell's controller, and the spell cannot resolve while the {4} is
// outstanding.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb5ed146-5079-46ef-a3e0-e584fa4467f6",
		Name:         "Spectral Interference",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target artifact or creature spell", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{4}",
				Question: "Spectral Interference — pay {4} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
