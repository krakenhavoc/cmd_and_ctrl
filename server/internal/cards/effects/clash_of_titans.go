package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clash of Titans — Instant {3}{R}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature fights another target creature. (Each deals damage
//	 equal to its power to the other.)"
//
// Blood Feud's two clauses and fight, at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "198c05ca-fb2d-4f15-a28f-1dad36466a4f",
		Name:         "Clash of Titans",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return slotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
