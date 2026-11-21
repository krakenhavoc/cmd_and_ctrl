package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "46665089-aa3d-44c3-964d-6638dfbb5782",
		Name:         "Essence Scatter",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
