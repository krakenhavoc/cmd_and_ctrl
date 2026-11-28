package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spore Cloud — Instant {1}{G}{G} (Fallen Empires, S82 OLD b855, #5637):
//
//	"Tap all blocking creatures. Prevent all combat damage that would be dealt this turn.
//	 Each attacking creature and each blocking creature doesn't untap during its
//	 controller's next untap step."
//
// In the printed order: the blockers are tapped, combat damage is prevented for the turn
// (Root Snare's prevention), and every creature attacking or blocking as it resolves sits
// out its controller's next untap step, tapped or not (Triton Tactics' hold).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36d55084-4696-4d16-ab3a-b00f1456603f",
		Name:         "Spore Cloud",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent | game.AnswerRestrict},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.TapEach(ctx, effects.BlockingCreatures(ctx.Game)); err != nil {
				return err
			}
			if err := (effects.PreventAllCombatDamageThisTurn{Label: "Spore Cloud: prevent combat damage"}).Apply(ctx); err != nil {
				return err
			}
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && (c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil) {
					ids = append(ids, c.InstanceID)
				}
			}
			return effects.DoesntUntapNextUntapStep{Targets: ids, Label: "Spore Cloud — doesn't untap during its controller's next untap step"}.Apply(ctx)
		},
	})
}
