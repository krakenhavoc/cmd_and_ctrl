package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valley Rannet — Creature — Beast {4}{R}{G}, 6/3 (Alara Reborn, #5586):
//
//	"Mountaincycling {2}, forestcycling {2} ({2}, Discard this card:
//	 Search your library for a Mountain or Forest card, reveal it, put it
//	 into your hand, then shuffle.)"
//
// The two landcycling keywords are two cycling abilities (CR 702.29e),
// each the shared Typecycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0a3b261-3daf-4082-8bc5-e66326a6c382",
		Name:         "Valley Rannet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
