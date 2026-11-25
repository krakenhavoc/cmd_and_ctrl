package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assemble the Rank and Vile — Conspiracy (Conspiracy: Take the Crown, S82
// OLD batch 268, #5540):
//
//	"Hidden agenda
//	 Creatures you control with the chosen name have "When this creature dies,
//	 you may pay {B}. If you do, create a tapped 2/2 black Zombie creature
//	 token.""
//
// Hidden agenda (ADR 0227). A layer-6 grant of a catalog bundle (ADR 0093):
// the granted dies trigger is the creature's own, so it is controlled by the
// creature's last controller (CR 603.10a). The token is made under that
// player's control, tapped, only if the {B} is really paid.
//
// No simplification.
func init() {
	const grantKey = "s82-old-b268/assemble-the-rank-and-vile-zombie"
	effects.Register(effects.Spec{
		OracleID:     "8c912b1f-4584-4945-9931-f6850e5ebfa4",
		Name:         "Assemble the Rank and Vile",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Assemble the Rank and Vile — you may pay {B}; if you do, create a tapped 2/2 black Zombie creature token",
					func(g *game.Game, item *game.StackItem) error {
						return effects.MayPay{
							Chooser:  item.Controller,
							Cost:     "{B}",
							Question: "Assemble the Rank and Vile — pay {B} to create a tapped 2/2 black Zombie creature token?",
							Then:     assembleTheRankAndVilePaid,
						}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "When this creature dies, you may pay {B}. If you do, create a tapped 2/2 black Zombie creature token.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(effects.CreaturesYouControlWithTheChosenName, grantKey)},
	})
}

var assembleTheRankAndVilePaid = effects.PayUnlessBranches("pay-unless/assemble-the-rank-and-vile-zombie",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("2/2 black Zombie")).EntersTapped(), N: 1}.Apply(ctx)
	}, nil)
