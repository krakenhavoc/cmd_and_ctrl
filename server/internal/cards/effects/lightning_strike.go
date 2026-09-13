package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Strike — Instant {1}{R} (EDHREC rank 2307):
//
//	"Lightning Strike deals 3 damage to any target."
//
// Lightning Bolt at two mana: the same any-target clause and the
// same three damage from the spell as its source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f34b9bc4-7bfe-47fd-ba23-4eeeb46026eb",
		Name:         "Lightning Strike",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 3}.Apply(ctx)
		},
	})
}
