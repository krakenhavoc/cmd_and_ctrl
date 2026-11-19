package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alchemist's Greeting — Sorcery {4}{R} (Innistrad Remastered):
//
//	"Alchemist's Greeting deals 4 damage to target creature.
//	 Madness {1}{R}"
func init() {
	Register(Spec{
		OracleID:     "9aaf0df7-5e2d-4f49-a8cb-66523be15ad6",
		Name:         "Alchemist's Greeting",
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		Completeness: CompletenessFull,
		Madness:      "{1}{R}",
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b33DamageChosenTargetFromSource(4)(ctx.Game, item)
		},
	})
}
