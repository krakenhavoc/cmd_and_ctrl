package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystical Dispute — Instant {2}{U}:
//
//	"This spell costs {2} less to cast if it targets a blue spell.
//	 Counter target spell unless its controller pays {3}."
//
// The reduction reads the announced target as the cost is determined (CR
// 601.2f), Ajani's Response's shape with a blue-spell test; the counter is
// the pay-unless prompt, asked of the spell's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48390b8d-db41-48a5-9bac-316660ab7252",
		Name:         "Mystical Dispute",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a blue spell.", OfColor("U")),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return counterTheTargetSpellUnlessPaid("Mystical Dispute", "{3}")(ctx.Game, item)
		},
	})
}
