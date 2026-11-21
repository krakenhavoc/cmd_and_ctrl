package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Igneous Pouncer — Creature — Elemental {4}{B}{R}, 5/1 (Alara Reborn,
// #5586):
//
//	"Haste
//	 Swampcycling {2}, mountaincycling {2} ({2}, Discard this card: Search
//	 your library for a Swamp or Mountain card, reveal it, put it into
//	 your hand, then shuffle.)"
//
// Haste is the printed keyword. The two landcycling keywords are two
// cycling abilities (CR 702.29e), each the shared Typecycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee6f1ddb-6749-4394-b2b0-4d293376f1af",
		Name:            "Igneous Pouncer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Swampcycling", "{2}", "a Swamp card", effects.IsLandWithSubtype("Swamp")),
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}
