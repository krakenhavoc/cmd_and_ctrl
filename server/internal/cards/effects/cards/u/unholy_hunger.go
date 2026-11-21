package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unholy Hunger — Instant {3}{B}{B} (Magic Origins, #5549):
//
//	"Destroy target creature.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, you gain 2 life."
//
// Murder, with spell mastery read at resolution (spellMastery), while the
// spell is still on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed8d8eb5-fbc8-4b7f-9b75-53ab118e72de",
		Name:         "Unholy Hunger",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyEachLegalTarget(item, ctx); err != nil {
				return err
			}
			if !effects.SpellMastery(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
