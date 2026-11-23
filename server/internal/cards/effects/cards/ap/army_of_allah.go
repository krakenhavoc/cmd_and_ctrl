package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3483946d-8645-4c22-b0ba-a65a44456324",
		Name:         "Army of Allah",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.AttackingCreature(),
				Power: 2,
				Label: "Army of Allah — attacking creatures get +2/+0",
			}.Apply(ctx)
		},
	})
}
