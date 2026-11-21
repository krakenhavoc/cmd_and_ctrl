package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Capricious Sliver — Creature — Sliver {3}{R}, 3/3 (Commander Masters):
//
//	"Sliver creatures you control have "Whenever this creature deals combat
//	 damage to a player, exile the top card of your library. You may play
//	 that card this turn.""
//
// A granted triggered ability (ADR 0093), Thorncaster Sliver's shape, on
// your own Slivers only. The body is the impulse bullet Breeches shares:
// exile the top card of the controller's library with a this-turn play
// permission (play, not cast, so a land may be played).
//
// No simplification.
const capriciousSliverGrant = "capricious-sliver/impulse"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f59b3d9-1535-4e13-a0a1-0f839fe90ac6",
		Name:         "Capricious Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: capriciousSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Capricious Sliver — exile the top card of your library; you may play it this turn",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						return effects.ExileTopCardYouMayPlayThisTurn(ctx.Item, ctx, 0)
					}))),
			},
			Text: "Whenever this creature deals combat damage to a player, exile the top card of your library. You may play that card this turn.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, capriciousSliverGrant),
		},
	})
}
