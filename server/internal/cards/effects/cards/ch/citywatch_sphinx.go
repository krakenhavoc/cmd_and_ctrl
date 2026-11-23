package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Citywatch Sphinx — Creature — Sphinx {5}{U}, 5/4 (Guilds of Ravnica, #5525):
//
//	"Flying
//	 When this creature dies, surveil 2."
//
// Flying is printed data; the dies trigger surveils 2 on the stack (CR
// 701.25a), as Failed Conversion's does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a63e0c27-cea8-408a-8e57-2f077826c50f",
		Name:            "Citywatch Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Citywatch Sphinx — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
