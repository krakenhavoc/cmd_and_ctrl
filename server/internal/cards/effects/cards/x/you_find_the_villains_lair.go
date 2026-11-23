package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You Find the Villains' Lair — Instant {1}{U}{U} (Adventures in the
// Forgotten Realms, #5499):
//
//	"Choose one —
//	 • Foil Their Scheme — Counter target spell.
//	 • Learn Their Secrets — Draw two cards, then discard two cards."
//
// Abjure's counter and Careful Study's loot, one per bullet. The draw
// resolves first and the discard is chosen from the hand after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41cdeb52-e90f-44f8-94d2-23aa032c4870",
		Name:         "You Find the Villains' Lair",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Foil Their Scheme — Counter target spell.",
				effects.TargetSpell("target spell"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if ctx.IsTargetLegal(t) {
							return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
						}
					}
					return nil
				}), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeDoing("Learn Their Secrets — Draw two cards, then discard two cards.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.LootOne(ctx.Game, item, 2)
				}),
		),
	})
}
