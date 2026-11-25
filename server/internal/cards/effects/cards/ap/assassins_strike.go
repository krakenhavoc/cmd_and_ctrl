package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assassin's Strike — Sorcery {4}{B}{B} (Jumpstart, #5510):
//
//	"Destroy target creature. Its controller discards a card."
//
// The controller is read before the destroy, so the discard happens even if
// the creature is indestructible or regenerates, and a creature gone before
// resolution means no spell effect at all (CR 608.2b). The controller chooses
// the card (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ad6339d-6206-49bf-bf97-05e6d3bcaab7",
		Name:         "Assassin's Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				controller := c.Controller
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if p := ctx.Game.PlayerByIDForEffect(controller); p == nil || p.Eliminated {
					return nil
				}
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: controller, Source: ctx.Source(), N: 1})
				return nil
			}
			return nil
		},
	})
}
