package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sage of Hours — Creature — Human Wizard {1}{U}, 1/1 (Secret Lair Drop):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put a
//	 +1/+1 counter on it.
//	 Remove all +1/+1 counters from this creature: For each five counters
//	 removed this way, take an extra turn after this one."
//
// The cost takes every +1/+1 counter (RemoveAllCountersFromThis) and the
// number removed (ctx.CountersRemoved) is the ability's own count; the
// extra turns are one TakeExtraTurn of that many (zero turns when fewer than
// five were removed, since TakeExtraTurn's zero N would mean one).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdbf70da-af38-4827-9b8b-39f553ebd34a",
		Name:         "Sage of Hours",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Sage of Hours — a +1/+1 counter on it (heroic)", effects.ThisGetsAPlusOneCounter),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove all +1/+1 counters from this creature: For each five counters removed this way, take an extra turn after this one.",
			Cost:    effects.RemoveAllCountersFromThis(game.CounterPlusOne),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := ctx.CountersRemoved() / 5
				if n <= 0 {
					return nil
				}
				return effects.TakeExtraTurn{Player: ctx.Controller(), N: n}.Apply(ctx)
			},
		}},
	})
}
