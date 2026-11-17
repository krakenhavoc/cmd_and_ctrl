package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emergency Weld — Sorcery {1}{B} (The Brothers' War, #5482):
//
//	"Return target artifact or creature card from your graveyard to your
//	 hand. Create a 1/1 colorless Soldier artifact creature token."
//
// Its one target is re-checked on resolution (CR 608.2b): with the card
// gone from the graveyard the spell does not resolve and no token is made.
// Otherwise the card returns and then the token is created.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86b0e978-5b4e-4bb7-bfda-be899c632d8e",
		Name:         "Emergency Weld",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target artifact or creature card from your graveyard", YouOwn(), Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := returnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			return CreateToken{Template: TokenCard(broB05SoldierToken), N: 1}.Apply(ctx)
		},
	})
}
