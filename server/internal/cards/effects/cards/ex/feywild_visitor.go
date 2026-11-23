package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feywild Visitor — Legendary Enchantment — Background {2}{U} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "Whenever one or more nontoken creatures
//	 you control deal combat damage to a player, you create a 1/1 blue Faerie
//	 Dragon creature token with flying.""
//
// A granted trigger (ADR 0093), the commander creature's own, with Professional
// Face-Breaker's batching: one trigger per player the controller's nontoken
// creatures connect with in a combat damage step (CR 603.2c), the creature read
// off the damage source after layers. Choose a Background is deck construction
// (ADR 0144).
//
// No simplification.
const feywildVisitorGrant = "feywild-visitor/faerie-dragon"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbc93912-2c36-4554-9b42-1dcca059fbb5",
		Name:         "Feywild Visitor",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: feywildVisitorGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.Not(effects.IsTokenPredicate()),
					"Feywild Visitor — create a 1/1 blue Faerie Dragon creature token with flying",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Faerie Dragon with flying"), N: 1})),
			},
			Text: "Whenever one or more nontoken creatures you control deal combat damage to a player, you create a 1/1 blue Faerie Dragon creature token with flying.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(feywildVisitorGrant)},
	})
}
