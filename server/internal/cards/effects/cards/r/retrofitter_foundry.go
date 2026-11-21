package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retrofitter Foundry — Artifact {1}:
//
//	"{3}: Untap this artifact.
//	 {2}, {T}: Create a 1/1 colorless Servo artifact creature token.
//	 {1}, {T}, Sacrifice a Servo: Create a 1/1 colorless Thopter artifact
//	 creature token with flying.
//	 {T}, Sacrifice a Thopter: Create a 4/4 colorless Construct artifact
//	 creature token."
//
// Four activated abilities, none timing-restricted. The untap is a plain
// untap of the Foundry itself; the other three are token makers whose
// sacrifice is a cost paid on announce (Pia and Kiran Nalaar's), so the
// token is made even if the Foundry is gone by resolution. Each {T} is the
// Foundry's own, and the sacrificed Servo or Thopter can be a token made
// by the previous ability the same turn: it has no {T} in the cost, so
// summoning sickness does not matter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaadfe41-b2be-4183-b45e-a70e53a59d2e",
		Name:         "Retrofitter Foundry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}: Untap this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item)) // #1432: the primitive knows "this"
				},
			},
			{
				Label:   "{2}, {T}: Create a 1/1 colorless Servo artifact creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1}),
			},
			{
				Label:   "{1}, {T}, Sacrifice a Servo: Create a 1/1 colorless Thopter artifact creature token with flying.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker | game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeN(1, "a Servo", effects.HasSubtype("Servo"))),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
			},
			{
				Label:   "{T}, Sacrifice a Thopter: Create a 4/4 colorless Construct artifact creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker | game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a Thopter", effects.HasSubtype("Thopter"))),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 colorless Construct artifact"), N: 1}),
			},
		},
	})
}
