package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spitting Spider — Creature — Spider {3}{G}{G}, 3/5 (Eighth Edition, S82 OLD
// batch 610, #5607):
//
//	"Reach
//	 Sacrifice a land: This creature deals 1 damage to each creature with
//	 flying."
//
// Reach is the engine's keyword. The sacrifice is the cost (Baloth Prime's
// B08SacrificeALand); the damage is Squall Line's each-creature-with-flying
// sweep without the players, flying read as the ability resolves, the Spider
// itself included if it has gained flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "24ddcf3c-a4f3-4373-9483-5ce6e50d6150",
		Name:            "Spitting Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label: "Sacrifice a land: This creature deals 1 damage to each creature with flying.",
			// ADR 0126 §6: only creatures with flying.
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.B08SacrificeALand(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.HasKeyword("flying")), 1)
			},
		}},
	})
}
