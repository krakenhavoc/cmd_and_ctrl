package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Insatiable Hemophage — Creature — Nightmare {3}{B}, 3/3 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {2}{B} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Deathtouch
//	 Whenever this creature mutates, each opponent loses X life and you gain X
//	 life, where X is the number of times this creature has mutated."
//
// Mutate is the engine's (ADR 0223). X is the merged permanent's count of
// mutations (S82OldB111TimesMutated), read as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "46bfdcec-7cdc-413e-aa77-ed98b64f147b",
		Name:             "Insatiable Hemophage",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		Purpose:          game.Purpose{LifeGain: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Insatiable Hemophage — each opponent loses X life and you gain X life, where X is the number of times this creature has mutated",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.S82OldB111XOfThisMutation(ctx)
					if err := (effects.EachOpponentLosesLife{N: x}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Amount: x}.Apply(ctx)
				}),
		},
	})
}
