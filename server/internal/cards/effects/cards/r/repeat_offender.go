package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repeat Offender — Creature — Human Assassin {1}{B}:
//
//	"{2}{B}: If this creature is suspected, put a +1/+1 counter on it.
//	 Otherwise, suspect it. (A suspected creature has menace and can't
//	 block.)"
//
// The branch is made as the ability RESOLVES, not at activation: two
// activations in a row, with the first still on the stack, suspect it
// and then grow it, which is the printed card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d1885d1-212e-47ae-865f-e6dc4148ddf2",
		Name:         "Repeat Offender",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: If this creature is suspected, put a +1/+1 counter on it. Otherwise, suspect it.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) { // #1432
					return nil
				}
				ctx := effects.NewContext(g, item)
				if g.IsSuspected(item.SourceCardID) {
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return effects.Suspect{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
