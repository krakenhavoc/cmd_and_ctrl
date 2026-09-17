package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tangle —
//
// "Prevent all combat damage that would be dealt this turn. Each attacking
// creature doesn't untap during its controller's next untap step."
func init() {
	Register(Spec{
		OracleID:     "f627e125-15af-4e53-b34e-82b60e4ec87b",
		Name:         "Tangle",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (PreventAllCombatDamageThisTurn{Label: "Tangle: prevent combat damage"}).Apply(ctx); err != nil {
				return err
			}
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.AttackingTarget != uuid.Nil {
					ids = append(ids, c.InstanceID)
				}
			}
			return (DoesntUntapNextUntapStep{Targets: ids, Label: "Tangle: attacking creatures don't untap"}).Apply(ctx)
		},
	})
}
