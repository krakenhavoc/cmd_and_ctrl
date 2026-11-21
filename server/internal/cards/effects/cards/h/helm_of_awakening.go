package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helm of Awakening — Artifact {2} (Dominaria Remastered):
//
//	"Spells cost {1} less to cast."
//
// Sphere of Resistance turned over: no predicate, so every spell every
// player casts costs {1} less, its controller's included. The reduction
// only takes away generic mana, never coloured (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2fcef65-4873-4b1f-a257-b8648097c990",
		Name:         "Helm of Awakening",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Spells cost {1} less to cast."),
		},
	})
}
