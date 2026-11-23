package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saber-Tooth Moose-Lion — Creature — Elk Cat {4}{G}{G}, 7/7:
//
//	"Reach
//	 Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Reach is printed. Forestcycling is typecycling (CR 702.29e) for any
// card with the Forest subtype, Rocksteady, Crash Courser's line.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7af493e3-15a1-42d6-a66c-82aacabe816f",
		Name:            "Saber-Tooth Moose-Lion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
