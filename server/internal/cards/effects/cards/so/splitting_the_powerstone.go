package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splitting the Powerstone — Sorcery {2}{U}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Create two tapped Powerstone tokens. If the sacrificed artifact was
//	 legendary, draw a card."
//
// "Was legendary" is read off the sacrificed artifact as it last existed
// on the battlefield (CR 608.2h), from the payment record (ADR 0113 §1).
// The Powerstones enter tapped because this effect says so; tapped is
// not part of the token's definition (the 2022-10-14 ruling).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "178828a5-c202-4503-9755-fc6bb2390209",
		Name:           "Splitting the Powerstone",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact", effects.Artifact()),
		// The draw waits for a prompt the tokens ask (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.CreateTokenAdvanced{
				Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(),
				N:    2,
			},
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.SacrificedHadSupertype(ctx, "Legendary") {
					return nil
				}
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
