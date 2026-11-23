package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treeshaker Chimera — Creature — Chimera, {5}{G}{G}, 8/5:
//
//	"All creatures able to block this creature do so.
//	 When this creature dies, draw three cards."
//
// #1684 leftover: the Lure clause is the Prized Unicorn shape
// (AllAbleToBlockDoSo); the dies trigger is WhenThisDies (graveyard
// only, CR 700.4) plus DrawCards. No new machinery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b119b69-4794-443e-aa13-747c125b5b1b",
		Name:         "Treeshaker Chimera",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Treeshaker Chimera — draw three cards", effects.Do(effects.DrawCards{N: 3})),
		},
	})
}
