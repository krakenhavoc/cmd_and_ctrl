package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Lindwurm — Creature — Wurm {4}{G}{G}, 6/6 (Kaldheim, #5505):
//
//	"When this creature enters, you gain 4 life."
//
// Migrating Ketradon's enters trigger without the reach and cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3583ec24-4f26-45ea-8fb4-a45bdd7bfc79",
		Name:         "Ravenous Lindwurm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ravenous Lindwurm — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
