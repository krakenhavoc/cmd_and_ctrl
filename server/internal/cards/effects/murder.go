package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murder — Instant {1}{B}{B} (EDHREC rank 843):
//
//	"Destroy target creature."
//
// The plainest removal spell in the format. Terminate's shape, one
// colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "938b4e2c-88d9-4637-bc00-e228920c9a78",
		Name:         "Murder",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
