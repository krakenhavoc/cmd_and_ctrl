package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golem Foundry — Artifact {3} (Foundations Commander):
//
//	"Whenever you cast an artifact spell, you may put a charge counter on
//	 this artifact.
//	 Remove three charge counters from this artifact: Create a 3/3
//	 colorless Golem artifact creature token."
//
// A "you may" cast trigger and a counter-removal cost (Magistrate's
// Scepter's RemoveCountersFromThis), paid at announce so a response
// cannot spend the counters twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e96c09f-c064-4dd8-85ee-0f48d86ab670",
		Name:         "Golem Foundry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.YouCast(effects.Artifact()), "Golem Foundry — you may put a charge counter on this artifact", effects.B33PutChargeCounterOnSelf),
				"Golem Foundry — put a charge counter on this artifact?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three charge counters from this artifact: Create a 3/3 colorless Golem artifact creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.RemoveCountersFromThis(game.CounterCharge, 3),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Golem artifact creature"), N: 1}),
		}},
	})
}
