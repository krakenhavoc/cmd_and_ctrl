package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamewright — Creature — Human Artificer {R}{W}, 1/1 (Conspiracy, S82 OLD
// batch 364, #5557):
//
//	"{1}, {T}: Create a 1/1 colorless Construct artifact creature token with
//	 defender.
//	 {T}, Sacrifice a creature with defender: This creature deals 1 damage to
//	 any target."
//
// Two activated abilities. The token is Sentinel Dispatch's 1/1 Construct with
// defender. The second sacrifices as a cost (CR 118.3), read through the
// layered view: a creature with defender, which may be a token Construct made
// by the first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c73f566-3b8a-4d57-b7a9-7437b6f1f66d",
		Name:         "Flamewright",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Create a 1/1 colorless Construct artifact creature token with defender.",
				Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Construct artifact with defender"), N: 1}),
			},
			{
				Label:   "{T}, Sacrifice a creature with defender: This creature deals 1 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a creature with defender", effects.Creature(), effects.HasKeyword("defender"))),
				Targets: effects.TargetAny(),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			},
		},
	})
}
