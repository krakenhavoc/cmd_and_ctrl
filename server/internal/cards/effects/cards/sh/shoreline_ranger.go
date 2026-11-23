package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shoreline Ranger — Creature — Bird Soldier Ranger {5}{U}, 3/4 (Scourge):
//
//	"Flying
//	 Islandcycling {2} ({2}, Discard this card: Search your library for an
//	 Island card, reveal it, put it into your hand, then shuffle.)"
//
// A printed keyword and the shared typecycling ability (CR 702.29e). The
// search finds any card with the Island subtype, a nonbasic Island
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "606fbf99-d3e3-4131-b991-3cb71c003834",
		Name:            "Shoreline Ranger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Islandcycling", "{2}", "an Island card", effects.IsLandWithSubtype("Island")),
		},
	})
}
