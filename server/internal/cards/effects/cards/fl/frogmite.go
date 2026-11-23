package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frogmite — Artifact Creature — Frog {4}, 2/2 (Mirrodin, #5606):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)"
//
// Affinity is the shared self cost modifier (CR 702.41a), generic mana
// only, so with four artifacts the Frogmite is free. The artifacts
// counted are permanents, so the Frogmite spell does not count itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a927fd33-f183-48d9-b582-55d379d70e13",
		Name:         "Frogmite",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
	})
}
