package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blasted Landscape — Land (Urza's Saga, #5624):
//
//	"{T}: Add {C}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The plain colourless tap (tapForColorless) and the shared CR 702.29a
// cycling ability from the hand. It enters untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9c8007ac-4b3d-4444-93e9-f583185e5d81",
		Name:          "Blasted Landscape",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
