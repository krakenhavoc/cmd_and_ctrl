package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Granitic Titan — Creature — Elemental {4}{R}{R}, 5/4
// (Hour of Devastation, #5534):
//
//	"Menace
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The keywords are printed data; cycling is the shared Cycling row, a
// hand ability that discards the card as its cost and draws (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d398868b-7574-40d5-8915-6393ee489067",
		Name:            "Granitic Titan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
