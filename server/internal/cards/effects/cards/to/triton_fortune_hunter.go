package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triton Fortune Hunter — Creature — Merfolk Soldier {2}{U}, 2/2 (Theros, #5561):
//
//	"Heroic — Whenever you cast a spell that targets this creature, draw a card."
//
// The shared heroic trigger (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f84637f0-1a04-4fd1-85d9-8e8f5d444800",
		Name:         "Triton Fortune Hunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Triton Fortune Hunter — draw a card (heroic)", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
