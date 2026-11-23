package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stampeding Scurryfoot — Creature — Mouse {G}, 1/1 (Aetherdrift, #4464):
//
//	"Exhaust — {3}{G}: Put a +1/+1 counter on this creature. Create a 3/3
//	 green Elephant creature token. (Activate each exhaust ability only
//	 once.)"
//
// Prowcatcher Specialist's exhaust (CR 702.177a: "Activate only once"),
// with a token after the counter. The two sentences are separate: if the
// Mouse has left the battlefield by resolution, the counter has nothing to
// land on (CR 608.2b) but the Elephant is still created.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b860806-0118-40f8-861b-225a7c1cf151",
		Name:         "Stampeding Scurryfoot",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Exhaust — {3}{G}: Put a +1/+1 counter on this creature. Create a 3/3 green Elephant creature token.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerMakesBlocker,
				SourceCounters: 1, Tokens: 1, CreatureTokens: 1},
			Exhaust: true,
			Cost:    effects.ManaCost("{3}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
					return err
				}
				return effects.CreateToken{Template: effects.TokenCard("3/3 green Elephant"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
