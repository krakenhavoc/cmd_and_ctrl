package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tarpan — Creature — Horse {G}, 1/1 (Ice Age, S82 OLD batch 851, #5636):
//
//	"When this creature dies, you gain 1 life."
//
// Anodet Lurker's death trigger for one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1e1ebc8-cd09-4280-a169-58ab7ece43ff",
		Name:         "Tarpan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tarpan — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
