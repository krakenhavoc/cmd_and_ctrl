package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spell Swindle — Instant {3}{U}{U}:
//
//	"Counter target spell. Create X Treasure tokens, where X is that
//	 spell's mana value."
//
// X is read off the spell before it is countered, while it is still on
// the stack (CR 608.2h, 202.3). A spell that can't be countered still
// pays out the Treasures; a spell that is gone by resolution makes the
// whole spell fizzle (CR 608.2b), as for any single-target spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81fa8e6a-5fb8-43de-802f-df5b985c6a7d",
		Name:         "Spell Swindle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			mv := 0
			if ctx.Game.StackItemForEffect(stackID) != nil {
				if card, ok := ctx.Game.LookupCardForEffect(stackID); ok {
					mv, _ = ctx.Game.ManaValueForEffect(card)
				}
			}
			if err := (effects.CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if mv <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: mv}.Apply(ctx)
		},
	})
}
