package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thopter Foundry — Artifact {W/B}{U} (Mystery Booster 2):
//
//	"{1}, Sacrifice a nontoken artifact: Create a 1/1 blue Thopter artifact
//	 creature token with flying. You gain 1 life."
//
// The sacrifice is a cost (it names a nontoken artifact, the Foundry itself
// included); the token and the life are the one effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88bef744-550e-4f33-b1ff-a8ee990ec754",
		Name:         "Thopter Foundry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a nontoken artifact: Create a 1/1 blue Thopter artifact creature token with flying. You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "a nontoken artifact", effects.Artifact(), effects.Not(effects.IsTokenPredicate()))),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Thopter artifact with flying"), N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
