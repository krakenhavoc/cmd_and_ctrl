package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Render Silent — Instant {W}{U}{U} (Dragon's Maze, S82 OLD b406, #5564):
//
//	"Counter target spell. Its controller can't cast spells this turn."
//
// Countersquall's counter (the spell's controller is read before the counter
// moves it) followed by Silence's cast ban on that player until end of turn
// (CR 101.2). The ban is made even when the spell can't be countered, since it
// is a second instruction and not a rider on the counter, and a spell that
// has already left the stack is not a legal target at resolution (CR 608.2b),
// so neither half happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3966bd03-a98e-43b6-a0c0-10d4d81c173c",
		Name:         "Render Silent",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				spell := ctx.Game.StackItemForEffect(t.ID)
				if spell == nil {
					return nil
				}
				controller := spell.Controller
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictCasting{
					Player:   controller,
					Rule:     game.CastBanRule{Kind: game.CastBanOutright},
					Label:    "Render Silent — can't cast spells this turn",
					Duration: effects.DurationUntilEndOfTurn(ctx),
				}.Apply(ctx)
			}
			return nil
		},
	})
}
