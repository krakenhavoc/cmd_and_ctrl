package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeletal Wurm — {7}{B} Creature — Skeleton Wurm 7/6 (Rise of the Eldrazi, S82 OLD #5581):
//
//	{B}: Regenerate this creature.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d988ed15-2927-421b-9a6e-01f0261272c2",
		Name:         "Skeletal Wurm",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}
