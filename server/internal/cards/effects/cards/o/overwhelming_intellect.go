package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overwhelming Intellect — Instant {4}{U}{U} (Duel Decks: Izzet vs. Golgari):
//
//	"Counter target creature spell. Draw cards equal to that spell's mana
//	 value."
//
// Counterpoint's read: the spell's mana value is taken while it is still on
// the stack (CR 202.3e, so an X spell counts its X), then it is countered, then
// you draw that many. A target that has left the stack by resolution makes the
// whole spell do nothing (CR 608.2b). A spell that can't be countered still
// leaves the draw: the counter is simply not applied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8038337b-aff9-49aa-800c-0e12d2688765",
		Name:         "Overwhelming Intellect",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			if ctx.Game.StackItemForEffect(stackID) == nil {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(stackID)
			if !ok {
				return nil
			}
			mv, _ := ctx.Game.ManaValueForEffect(card)
			if err := (effects.CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: mv}.Apply(ctx)
		},
	})
}
