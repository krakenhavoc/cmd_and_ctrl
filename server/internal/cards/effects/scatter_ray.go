package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scatter Ray — Instant for {1}{U} (The Brothers' War):
//
//	"Counter target artifact or creature spell unless its controller pays {4}."
func init() {
	Register(Spec{
		OracleID:     "b515dfc6-1a1e-46d6-a7a3-0e3af8279f49",
		Name:         "Scatter Ray",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target artifact or creature spell", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{4}",
				Question: "Scatter Ray — pay {4}, or your spell is countered?",
			}.Apply(ctx)
		},
	})
}
