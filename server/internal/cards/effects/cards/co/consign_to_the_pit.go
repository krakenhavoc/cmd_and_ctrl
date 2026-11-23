package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consign to the Pit — Sorcery {5}{B} (Jumpstart 2022, #5948):
//
//	"Destroy target creature. Consign to the Pit deals 2 damage to that
//	 creature's controller."
//
// The controller is read before the destroy (CR 608.2h), so a creature
// that is regenerated or indestructible still has its controller hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0c50079-5376-47ff-82c5-d52dbf49afdf",
		Name:         "Consign to the Pit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found {
				return nil
			}
			controller := c.Controller
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
