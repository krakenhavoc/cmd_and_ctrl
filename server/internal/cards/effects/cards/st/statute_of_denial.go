package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Statute of Denial — Instant {2}{U}{U} (Magic 2015, #5556):
//
//	"Counter target spell. If you control a blue creature, draw a card, then
//	 discard a card."
//
// Ionize's counter. The loot is the second sentence, and it happens whether
// or not the counter did anything: a target that is gone leaves the whole
// spell countered on resolution (CR 608.2b), so nothing follows then. The
// blue creature is read as the spell resolves, through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03e04cee-4727-44fe-96f3-42cbd08beb76",
		Name:         "Statute of Denial",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			if !effects.S82OldB357YouControlABlueCreature(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), item.SourceCardID, 1, 1, "")
		},
	})
}
