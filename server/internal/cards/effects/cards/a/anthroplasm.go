package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anthroplasm — Creature — Shapeshifter {2}{U}{U} 0/0:
//
//	"This creature enters with two +1/+1 counters on it.
//	 {X}, {T}: Remove all +1/+1 counters from this creature and put X +1/+1
//	 counters on it."
//
// The removal is part of the effect, not a cost, so X is paid and the counters
// come off as the ability resolves (and only if it is still this creature).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7536d94-4062-4ccc-a086-1c47954b24c7",
		Name:         "Anthroplasm",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("+1/+1", 2, "Anthroplasm: enters with two +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Remove all +1/+1 counters from this creature and put X +1/+1 counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := effects.RemoveAllCountersFromThisEffect(game.CounterPlusOne)(g, item); err != nil {
					return err
				}
				if ctx.X() <= 0 {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
