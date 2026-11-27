package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Precinct Captain — Creature — Human Soldier {W}{W}, 2/2 (Return to Ravnica,
// #5566):
//
//	"First strike
//	 Whenever this creature deals combat damage to a player, create a 1/1
//	 white Soldier creature token."
//
// Moonblade Shinobi's self-only combat damage trigger with the 1/1 white
// Soldier token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3c75234b-e367-46da-8454-be4315f9b9de",
		Name:            "Precinct Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Precinct Captain — create a 1/1 white Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
