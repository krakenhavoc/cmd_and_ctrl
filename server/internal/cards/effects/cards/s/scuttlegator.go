package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scuttlegator — Creature — Crab Turtle Crocodile {4}{G/U}{G/U}, 6/6
// (Ravnica: Clue Edition):
//
//	"Defender
//	 {6}{G/U}{G/U}: Adapt 3. (If this creature has no +1/+1 counters on it,
//	 put three +1/+1 counters on it.)
//	 As long as this creature has a +1/+1 counter on it, it can attack as
//	 though it didn't have defender."
//
// Demon Wall's defender waiver, keyed to +1/+1 counters only, with
// Pteramander's adapt (adaptN).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9bc0ca04-26ab-4743-bc65-0b5f4115c8f7",
		Name:            "Scuttlegator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			effects.AvB18CanAttackAsThoughNoDefenderWhile(
				"As long as this creature has a +1/+1 counter on it, it can attack as though it didn't have defender.",
				func(_ *game.Game, source *game.Card) bool {
					return source != nil && source.Counters[game.CounterPlusOne] > 0
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}{G/U}{G/U}: Adapt 3.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{6}{G/U}{G/U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 3) },
		}},
	})
}
