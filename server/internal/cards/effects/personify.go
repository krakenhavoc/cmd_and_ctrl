package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Personify — Instant {1}{W} (Lorwyn Eclipsed, #4458):
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control. Create a 1/1 colorless
//	 Shapeshifter creature token with changeling. (It's every creature
//	 type.)"
//
// Ephemerate's flicker (flickerTheTargetToItsOwner): the card returns as
// a new object under its owner's control (CR 400.7), without its counters
// or Auras; a token exiled this way ceases to exist and doesn't return
// (CR 111.8, the card's ruling). Then the spell creates the printed
// Shapeshifter token. If the target is illegal as Personify resolves, the
// spell does nothing at all, the token included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1708ad9-35d4-4243-92ca-ddf9d129aa32",
		Name:         "Personify",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := flickerTheTargetToItsOwner(item, ctx); err != nil {
				return err
			}
			return CreateToken{Template: TokenCard("1/1 colorless Shapeshifter with changeling"), N: 1}.Apply(ctx)
		},
	})
}
