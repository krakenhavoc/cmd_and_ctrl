package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pit Fight — Instant {1}{R/G} (Ravnica: Clue Edition, #5929):
//
//	"Target creature you control fights another target creature."
//
// Two target clauses, the second Distinct from the first ("another"). The
// fight is slotZeroFightsSlotOne: nothing fights once either creature has
// left or become illegal (CR 701.14b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef4b0ced-70a1-41d8-9aed-34e75f979c5f",
		Name:         "Pit Fight",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(TargetCreature("another target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return slotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
