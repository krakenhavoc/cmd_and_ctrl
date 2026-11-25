package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skeletal Vampire — Creature — Vampire Skeleton {4}{B}{B}, 3/3 (Commander
// 2017, #5533):
//
//	"Flying
//	 When this creature enters, create two 1/1 black Bat creature tokens
//	 with flying.
//	 {3}{B}{B}, Sacrifice a Bat: Create two 1/1 black Bat creature tokens
//	 with flying.
//	 Sacrifice a Bat: Regenerate this creature."
//
// Both Bat abilities sacrifice as a cost (CR 118.3), so the Bat is gone before
// the ability goes on the stack and cannot be responded to by saving it. The
// regeneration is a shield on the Vampire (CR 701.19a) created as the ability
// resolves. Any Bat may be sacrificed, a token or a card, one a Bat only
// through a type-changing effect included.
//
// No simplification.
func init() {
	sacBat := game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Bat", effects.HasSubtype("Bat"))}
	effects.Register(effects.Spec{
		OracleID:        "7ed8a69b-fba0-4d4d-98b2-311a92373669",
		Name:            "Skeletal Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skeletal Vampire — create two 1/1 black Bat creature tokens with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Bat with flying"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{B}{B}, Sacrifice a Bat: Create two 1/1 black Bat creature tokens with flying.",
				Purpose: game.Purpose{Tokens: 2, Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{3}{B}{B}"), sacBat),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 black Bat with flying"), N: 2}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "Sacrifice a Bat: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
				Cost:    sacBat,
				Effect:  effects.RegenerateThisSource,
			},
		},
	})
}
