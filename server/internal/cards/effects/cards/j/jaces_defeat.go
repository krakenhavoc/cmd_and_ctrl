package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace's Defeat — {1}{U} Instant (Hour of Devastation):
//
//	"Counter target blue spell. If it was a Jace planeswalker spell, scry 2."
//
// The spell is read as it last existed before it is countered (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ed8e27b-2137-4e19-99bb-5e0d6d32e3af",
		Name:         "Jace's Defeat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      effects.TargetSpell("target blue spell", effects.OfColor("U")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			wasJace := false
			if len(item.Targets) > 0 {
				if c, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID); ok {
					wasJace = c.IsPlaneswalker() && c.HasSubtype("Jace")
				}
			}
			if err := effects.CounterTheTargetSpell(item, ctx); err != nil {
				return err
			}
			if !wasJace {
				return nil
			}
			return effects.Scry{N: 2}.Apply(ctx)
		},
	})
}
