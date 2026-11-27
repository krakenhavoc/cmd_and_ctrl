package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Metathran Zombie (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Metathran Zombie {1}{U}, 1/1:
//	"{B}: Regenerate this creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c631887c-e157-4a95-8a59-08a44854d07f",
		Name:         "Metathran Zombie",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}
