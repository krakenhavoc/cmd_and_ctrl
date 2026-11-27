package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeletal Kathari — Creature — Bird Skeleton {4}{B}, 3/2 (Shards of Alara,
// S82 OLD batch 510, #5588):
//
//	"Flying
//	 {B}, Sacrifice a creature: Regenerate this creature."
//
// The sacrificed creature may be the Kathari itself, as printed.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5bcb63b4-6236-45c5-884d-188f8f5e8f02",
		Name:            "Skeletal Kathari",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB510RegenerateRow("{B}, Sacrifice a creature: Regenerate this creature.",
				effects.Plus(effects.ManaCost("{B}"), effects.SacrificeACreature())),
		},
	})
}
