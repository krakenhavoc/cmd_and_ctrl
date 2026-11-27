package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmering Barrier — Creature — Wall 1/3 {1}{W} (Urza's Saga, S82 OLD b722):
//
//	Defender
//	First strike
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Defender and first strike are printed keywords; cycling is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94674840-c73c-4659-80b9-11140d1deaa3",
		Name:            "Shimmering Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "first strike"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
