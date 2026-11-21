package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skatewing Spy — Creature — Vedalken Rogue Mutant {3}{U}, 2/3 (Ravnica
// Allegiance, #5521):
//
//	"{5}{U}: Adapt 2. (If this creature has no +1/+1 counters on it, put
//	 two +1/+1 counters on it.)
//	 Each creature you control with a +1/+1 counter on it has flying."
//
// Adapt is adaptN (CR 701.46a): the counters go on only if it has no +1/+1
// counter as the ability resolves. The flying grant is Abzan Falconer's
// layer-6 static over creatures you control with a +1/+1 counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49701c15-c6f7-4b72-9e89-d34148e5c675",
		Name:         "Skatewing Spy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 2) },
		}},
		Static: []game.StaticAbility{
			effects.CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "flying"),
		},
	})
}
