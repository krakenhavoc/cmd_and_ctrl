package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "178828a5-c202-4503-9755-fc6bb2390209",
		Name:           "Splitting the Powerstone",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact", Artifact()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CreateTokenAdvanced{
				Controller: ctx.Controller(),
				Spec:       Token(PowerstoneToken()).EntersTapped(),
				N:          2,
			}).Apply(ctx); err != nil {
				return err
			}
			if !sacrificedHadSupertype(ctx, "Legendary") {
				return nil
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
