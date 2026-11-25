package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Youthful Scholar — Creature — Human Wizard {3}{U}, 2/2 (Dragons of Tarkir, S82 old
// batch 329):
//
//	"When this creature dies, draw two cards."
//
// Darkslick Drake's dies trigger drawing two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "070bd1e8-6c61-4b0e-b179-e375b843410f",
		Name:         "Youthful Scholar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Youthful Scholar — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
