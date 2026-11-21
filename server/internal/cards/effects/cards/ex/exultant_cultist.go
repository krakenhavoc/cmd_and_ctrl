package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exultant Cultist — Creature — Human Wizard {2}{U}, 2/2 (Eldritch Moon,
// #5541):
//
//	"When this creature dies, draw a card."
//
// Spore Crawler's dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e802442e-6a78-4c15-9034-fc8872470208",
		Name:         "Exultant Cultist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Exultant Cultist — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
