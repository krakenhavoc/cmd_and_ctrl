package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Army of Allah — Instant {1}{W}{W} (Arabian Nights, #5641):
//
//	"Attacking creatures get +2/+0 until end of turn."
//
// Auron's Inspiration without the flashback. The set of attacking
// creatures is locked in as the spell resolves: every attacking
// creature, whoever controls it, gets the bonus, and a creature that
// starts attacking later in the turn does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3483946d-8645-4c22-b0ba-a65a44456324",
		Name:         "Army of Allah",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match: AttackingCreature(),
				Power: 2,
				Label: "Army of Allah — attacking creatures get +2/+0",
			}.Apply(ctx)
		},
	})
}
