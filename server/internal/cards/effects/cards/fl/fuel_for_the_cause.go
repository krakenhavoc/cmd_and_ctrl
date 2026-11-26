package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fuel for the Cause — Instant {2}{U}{U} (Mirrodin Besieged, S82 OLD b467,
// #5576):
//
//	"Counter target spell, then proliferate."
//
// Two instructions in printed order (CR 608.2c). The proliferate (CR 701.34)
// happens whether or not the counter did anything, as long as the spell
// resolves at all; a target spell gone by then makes the whole spell fizzle
// (CR 608.2b), so nothing proliferates.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "667cffdd-8951-4223-8895-fa46e9f14b96",
		Name:         "Fuel for the Cause",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			// The clause is judged before the counter: a countered spell has
			// left the stack, which would read as an illegal target after.
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			if err := effects.SomB1CounterTheTargetSpells(item, ctx); err != nil {
				return err
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
