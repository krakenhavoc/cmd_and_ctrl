package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essence Scatter — Instant {1}{U} (EDHREC rank 2708):
//
//	"Counter target creature spell."
//
// Negate's other half. The clause is enforced at announce and again
// at resolution: the picker offers only creature spells, and the
// engine refuses a noncreature pick (CR 601.2c). An artifact
// creature or an enchantment creature is a creature spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46665089-aa3d-44c3-964d-6638dfbb5782",
		Name:         "Essence Scatter",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
