package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Popular Entertainer — Legendary Enchantment — Background {1}{R} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "Whenever one or more creatures you
//	 control deal combat damage to a player, goad target creature that player
//	 controls.""
//
// A granted trigger (ADR 0093), the commander creature's own: once for each
// player the creatures its controller controls connect with in a combat
// damage step, however many connect (CR 603.2c). The target is a creature
// controlled by the player who was dealt the damage (ControlledByThatPlayer),
// chosen as the trigger goes on the stack; goad (CR 701.15) is Kang Dynasty's,
// lasting until its controller's next turn. Choose a Background is deck
// construction (ADR 0144).
//
// No simplification.
const popularEntertainerGrant = "popular-entertainer/goad"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "702d7b7b-7935-4cb2-9cd7-06f638f5405f",
		Name:         "Popular Entertainer",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: popularEntertainerGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
					"Popular Entertainer — goad target creature that player controls",
					effects.GalvanizingTenorGoad),
					effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls"))),
			},
			Text: "Whenever one or more creatures you control deal combat damage to a player, goad target creature that player controls.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(popularEntertainerGrant)},
	})
}
