package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "86b0e978-5b4e-4bb7-bfda-be899c632d8e",
		Name:         "Emergency Weld",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target artifact or creature card from your graveyard", effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.ReturnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TokenCard(effects.BroB05SoldierToken), N: 1}.Apply(ctx)
		},
	})
}
