package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spore Crawler — Creature — Fungus {2}{G}, 3/2 (Innistrad: Crimson Vow):
//
//	"When this creature dies, draw a card."
//
// Buzz Bots' dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa7c13c0-350b-4ed6-8c16-590d2b988184",
		Name:         "Spore Crawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Spore Crawler — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
