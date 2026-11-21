package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blur of Blades — Instant {1}{R} (Hour of Devastation, #5534):
//
//	"Put a -1/-1 counter on target creature. Blur of Blades deals 2 damage
//	 to that creature's controller."
//
// One target. The counter goes on first; the damage goes to the creature's
// controller as the spell resolves, and is dealt whether or not the
// counter could be put on it. A target that left in response does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65410f7a-c749-4b23-ad61-8a7136efcad2",
		Name:         "Blur of Blades",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := c.Controller
			if err := (effects.AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
