package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marchesa's Infiltrator — Creature — Human Rogue {2}{U}, 1/1 (Conspiracy,
// S82 OLD batch 363, #5557):
//
//	"Dethrone (Whenever this creature attacks the player with the most life
//	 or tied for most life, put a +1/+1 counter on it.)
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// Dethrone is the shared GapB201Dethrone trigger (CR 702.105a). The draw is
// the controller's, on the creature's combat damage to a player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f05fab1-398a-4bd1-824d-6b1373dc9a94",
		Name:         "Marchesa's Infiltrator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GapB201Dethrone("Marchesa's Infiltrator — dethrone: put a +1/+1 counter on it"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Marchesa's Infiltrator — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
