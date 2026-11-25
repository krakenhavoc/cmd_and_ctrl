package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nine-Tail White Fox (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	Whenever this creature deals combat damage to a player, draw a card.
//
// {2}{U} Creature — Fox Spirit 2/2. A mandatory draw (Broodbirth Viper without the "may").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51f4eb1f-abf2-425a-b78e-cafac177da89",
		Name:         "Nine-Tail White Fox",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Nine-Tail White Fox — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
