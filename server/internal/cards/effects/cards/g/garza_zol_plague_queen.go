package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garza Zol, Plague Queen — Legendary Creature — Vampire Noble {4}{U}{B}{R},
// 5/5 (Coldsnap, #5597):
//
//	"Flying, haste
//	 Whenever a creature dealt damage by Garza Zol this turn dies, put a
//	 +1/+1 counter on Garza Zol.
//	 Whenever Garza Zol deals combat damage to a player, you may draw a card."
//
// Sengir Vampire's counter trigger and Shadowmage Infiltrator's optional draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b48e926-fc0e-403c-a47b-1f1d96e2eaf7",
		Name:            "Garza Zol, Plague Queen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Garza Zol — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.OptionalDraw(effects.WheneverThisDealsCombatDamageToAPlayer("Garza Zol — you may draw a card",
				effects.Do(effects.DrawCards{N: 1})), "Draw a card?", 1),
		},
	})
}
