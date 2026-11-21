package cardssy

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tangle —
//
// "Prevent all combat damage that would be dealt this turn. Each attacking
// creature doesn't untap during its controller's next untap step."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f627e125-15af-4e53-b34e-82b60e4ec87b",
		Name:         "Tangle",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.PreventAllCombatDamageThisTurn{Label: "Tangle: prevent combat damage"}).Apply(ctx); err != nil {
				return err
			}
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.AttackingTarget != uuid.Nil {
					ids = append(ids, c.InstanceID)
				}
			}
			return (effects.DoesntUntapNextUntapStep{Targets: ids, Label: "Tangle: attacking creatures don't untap"}).Apply(ctx)
		},
	})
}
