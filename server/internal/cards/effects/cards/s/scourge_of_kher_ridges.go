package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scourge of Kher Ridges — Creature — Dragon {6}{R}{R}, 6/6
// (Future Sight, #5594):
//
//	"Flying
//	 {1}{R}: This creature deals 2 damage to each creature without
//	 flying.
//	 {5}{R}: This creature deals 6 damage to each other creature with
//	 flying."
//
// Harbinger of the Hunt's two sweeps at Scourge's amounts. Flying is read
// through the layers as each sweep resolves, the first sweep spares the
// Dragon itself because it flies, and the second excludes it by identity
// ("each other").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a3ced75d-eb1b-4a56-bbe0-49e8b906f729",
		Name:            "Scourge of Kher Ridges",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}: This creature deals 2 damage to each creature without flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
				Cost:    effects.ManaCost("{1}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 2)
				},
			},
			{
				Label:   "{5}{R}: This creature deals 6 damage to each other creature with flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6, Partial: true}},
				Cost:    effects.ManaCost("{5}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.HasKeyword("flying"),
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != source }), 6)
				},
			},
		},
	})
}
