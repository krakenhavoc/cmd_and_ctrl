package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timberland Ancient — Creature — Treefolk {4}{G}{G}, 6/5 (March of the
// Machine):
//
//	"Reach, trample
//	 Forestcycling {2} ({2}, Discard this card: Search your library for
//	 a Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Forestcycling is the shared Typecycling constructor (CR 702.29e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e71eca6f-388c-4750-b70e-3ee1832931f0",
		Name:            "Timberland Ancient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
