package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ed8d8eb5-fbc8-4b7f-9b75-53ab118e72de",
		Name:         "Unholy Hunger",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyEachLegalTarget(item, ctx); err != nil {
				return err
			}
			if !spellMastery(ctx.Game, ctx.Controller()) {
				return nil
			}
			return GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
