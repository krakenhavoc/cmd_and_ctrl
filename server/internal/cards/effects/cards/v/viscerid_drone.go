package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Viscerid Drone — Creature — Homarid Drone {1}{U} 1/2 (Alliances, S82 OLD
// batch 820):
//
//	"{T}, Sacrifice a creature and a Swamp: Destroy target nonartifact
//	 creature. It can't be regenerated.
//	 {T}, Sacrifice a creature and a snow Swamp: Destroy target creature. It
//	 can't be regenerated."
//
// Both are one sacrifice of two different permanents (a creature and a Swamp
// are matched one to one, so an animated Swamp pays only one half), and the
// snow version names the snow supertype (game.SacrificeKind.Supertypes).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fd4be4c-70a3-4c91-a4c6-788a2051ff13",
		Name:         "Viscerid Drone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice a creature and a Swamp: Destroy target nonartifact creature. It can't be regenerated.",
				Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
				Cost: effects.Plus(effects.TapCost(), effects.SacrificeEach("a creature and a Swamp",
					effects.SacrificeCardType("a creature", "creature"), effects.SacrificeSubtype("a Swamp", "Swamp"))),
				Targets: effects.TargetCreature("target nonartifact creature", effects.Not(effects.Artifact())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "{T}, Sacrifice a creature and a snow Swamp: Destroy target creature. It can't be regenerated.",
				Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
				Cost: effects.Plus(effects.TapCost(), effects.SacrificeEach("a creature and a snow Swamp",
					effects.SacrificeCardType("a creature", "creature"), effects.OldB820SacrificeSupertypeSubtype("a snow Swamp", "snow", "Swamp"))),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
				},
			},
		},
	})
}
