package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mu Yanling, Wind Rider — Legendary Creature — Human Wizard Pilot
// {2}{U}{U}, 2/4:
//
//	"When Mu Yanling enters, create a 3/2 colorless Vehicle artifact
//	 token with crew 1.
//	 Vehicles you control have flying.
//	 Whenever one or more creatures you control with flying deal combat
//	 damage to a player, draw a card."
//
// The token is a Vehicle token with its own crew ability
// (VehicleCrew1Token, #3767). "Vehicles you control have flying" is a
// layer-6 grant to every Vehicle you control, crewed or not, so a
// crewed Vehicle attacks with flying. The draw is one trigger per
// player your flying creatures connect with in a combat damage step
// (CR 603.2c), reading flying as each creature deals its damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaafcd69-6295-4d79-8e67-98df0c9bae88",
		Name:         "Mu Yanling, Wind Rider",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype("Vehicle")
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mu Yanling, Wind Rider — create a 3/2 Vehicle token with crew 1", effects.Do(effects.CreateToken{Template: effects.VehicleCrew1Token(), N: 1})),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.HasKeyword("flying"),
				"Mu Yanling, Wind Rider — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
