package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Ravager — Artifact Creature — Beast {2}, 0/0 (EDHREC rank
// 3694):
//
//	"Sacrifice an artifact: Put a +1/+1 counter on this creature.
//	 Modular 1"
//
// Modular is the engine's keyword (#2012, game/modular.go): it enters
// with its +1/+1 counter, and when it dies its counters may move to
// target artifact creature. The Ravager may sacrifice itself; the
// counter then lands nowhere (it is in the graveyard), and its modular
// trigger moves the counters it had.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "62e7e7b1-9887-4d15-b0e5-a8ddc711bd88",
		Name:            "Arcbound Ravager",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 1"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an artifact", effects.Artifact())},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			},
		}},
	})
}
