package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trollbred Guardian — Creature — Troll Frog Warrior {4}{G}, 5/5 (Ravnica
// Allegiance, #5521):
//
//	"{2}{G}: Adapt 2. (If this creature has no +1/+1 counters on it, put
//	 two +1/+1 counters on it.)
//	 Each creature you control with a +1/+1 counter on it has trample."
//
// Adapt is adaptN (CR 701.46a): the counters go on only if it has no +1/+1
// counter as the ability resolves. The trample grant is the shared layer-6
// static over creatures you control with a +1/+1 counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9c756a5-f4d4-4b84-be60-ad4bc0d1acf9",
		Name:         "Trollbred Guardian",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 2) },
		}},
		Static: []game.StaticAbility{
			effects.CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "trample"),
		},
	})
}
