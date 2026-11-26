package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jungle Weaver — Creature — Spider {5}{G}{G}, 5/6 (Shards of Alara, S82 OLD
// batch 509, #5588):
//
//	"Reach
//	 Cycling {2}"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a585186c-d7c2-4f8b-acdc-c5f9ab950398",
		Name:            "Jungle Weaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
