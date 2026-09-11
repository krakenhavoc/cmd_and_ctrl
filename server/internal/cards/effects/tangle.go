package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangle — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt this turn."
//	"Each attacking creature doesn't untap during its controller's
//	 next untap step."
//
// DECLARED SIMPLIFICATION — the second line is a no-op. "Doesn't
// untap during its controller's next untap step" is a CR 614-style
// continuous effect on a step the engine performs unconditionally:
// the untap step untaps every permanent its controller owns, with
// no per-permanent exception and no "skipped untaps" marker on
// Card. Nothing exists to write the exception into.
//
// That makes Tangle strictly a two-mana Fog today — weaker than
// printed, never stronger, which is the acceptable direction. It is
// declared here rather than omitted because the missing half is
// exactly one engine feature away (a skip-next-untap flag on Card
// plus a read in the untap step), and a card file that says so is
// how the next person finds the ticket. The fog half is the half
// that gets cast in Commander: Tangle is played as a green Fog that
// happens to have a rider.
func init() {
	Register(Spec{
		OracleID: "f627e125-15af-4e53-b34e-82b60e4ec87b",
		Name:     "Tangle",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PreventAllCombatDamageThisTurn{
				Label: "Tangle: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
