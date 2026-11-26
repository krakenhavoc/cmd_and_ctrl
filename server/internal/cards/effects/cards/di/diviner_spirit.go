package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diviner Spirit — Creature — Spirit {4}{U}, 2/4 (Commander 2013, S82 OLD b380,
// #5560):
//
//	"Whenever this creature deals combat damage to a player, you and that player
//	 each draw that many cards."
//
// Black Widow's "you and that player each draw", over any player the Spirit
// hit and the combat damage dealt (S82OldB380DivinerSpiritDraw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "911b8849-dd0a-4383-8403-ea80227c5d7d",
		Name:         "Diviner Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Diviner Spirit — you and that player each draw that many cards", effects.S82OldB380DivinerSpiritDraw),
		},
	})
}
