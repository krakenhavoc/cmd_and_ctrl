package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shortcut Seeker — Creature — Human Rogue {3}{U}, 2/5:
//
//	"Whenever this creature deals combat damage to a player, venture
//	 into the dungeon."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27cfa916-282d-476c-8ee9-ca7f0d84f551",
		Name:         "Shortcut Seeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(effects.VentureLabel("Shortcut Seeker"), effects.Venture),
		},
	})
}
