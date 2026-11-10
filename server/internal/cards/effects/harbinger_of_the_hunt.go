package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harbinger of the Hunt — Creature — Dragon {3}{R}{G}, 5/3:
//
//	"Flying
//	 {2}{R}: This creature deals 1 damage to each creature without flying.
//	 {2}{G}: This creature deals 1 damage to each other creature with
//	 flying."
//
// Two activated abilities, each one batch of damage from the Harbinger
// (damageEachMatching). The first has no "other", so a Harbinger that has
// lost flying hits itself (ruling); the second excludes the Harbinger by
// identity. Flying is read through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3256029f-6558-4fcc-9fa1-74ceba3e5c92",
		Name:            "Harbinger of the Hunt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{R}: This creature deals 1 damage to each creature without flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
				Cost:    ManaCost("{2}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item), And(Creature(), WithoutKeyword("flying")), 1)
				},
			},
			{
				Label:   "{2}{G}: This creature deals 1 damage to each other creature with flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
				Cost:    ManaCost("{2}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					return damageEachMatching(NewContext(g, item), And(Creature(), HasKeyword("flying"),
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != source }), 1)
				},
			},
		},
	})
}
