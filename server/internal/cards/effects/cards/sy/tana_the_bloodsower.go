package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tana, the Bloodsower — Legendary Creature — Elf Druid {2}{R}{G}, 2/2
// (Commander 2016, #5537):
//
//	"Trample
//	 Whenever Tana deals combat damage to a player, create that many 1/1
//	 green Saproling creature tokens.
//	 Partner (You can have two commanders if both have partner.)"
//
// Lathril's "that many", read off the damage event, so only the damage
// that reached the player counts (trample damage assigned past a blocker
// included, CR 702.19c), with the Saproling token. Partner is a
// deck-building rule (CR 702.124) that internal/deck reads off the
// oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7dff624-1f6b-4e00-b13b-0b0d69a48ec2",
		Name:            "Tana, the Bloodsower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Tana, the Bloodsower — create that many 1/1 green Saproling tokens",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("1/1 green Saproling"),
					Name:     "the combat damage dealt",
					Count:    func(ctx *effects.Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
	})
}
