package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spell Snuff — Instant {1}{U}{U} (Modern Horizons):
//
//	"Counter target spell.
//	 Fateful hour — If you have 5 or less life, draw a card."
//
// Counterspell's counter, then Courageous Resolve's fateful hour: your life
// total is read as the spell resolves (CR 608.2c), so a hit the countered spell
// would have dealt does not matter. The draw is part of the same resolution
// whether or not the target was still there to counter, but with the target
// gone the whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19f3ec61-f334-4708-ad93-04ebe15c1798",
		Name:         "Spell Snuff",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			p := ctx.PlayerByID(ctx.Controller())
			if p == nil || p.Life > 5 {
				return nil
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
