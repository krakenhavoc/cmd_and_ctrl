package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neurok Commando — Creature — Human Rogue {1}{U}{U}, 2/1 (Mirrodin Besieged,
// #5576):
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Whenever this creature deals combat damage to a player, you may
//	 draw a card."
//
// Shroud is a printed keyword (CR 702.18a); the draw is Shadowmage
// Infiltrator's optional combat-damage trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48936751-429b-4c40-88f5-72acde6278cc",
		Name:            "Neurok Commando",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.WheneverThisDealsCombatDamageToAPlayer("Neurok Commando — you may draw a card",
				effects.Do(effects.DrawCards{N: 1})), "Draw a card?", 1),
		},
	})
}
