package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chittering Harvester — Creature — Nightmare {5}{B}, 4/6 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {4}{B} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Whenever this creature mutates, each opponent sacrifices a creature of
//	 their choice."
//
// Mutate is the engine's (ADR 0223). Rot-Tide Gargantua's edict: each
// opponent, in turn order, chooses and sacrifices one creature; one with no
// creature sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9473b445-294b-4572-8267-f1f2fc2a05a7",
		Name:             "Chittering Harvester",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{4}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, "Chittering Harvester — each opponent sacrifices a creature of their choice",
				func(g *game.Game, item *game.StackItem) error {
					return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
