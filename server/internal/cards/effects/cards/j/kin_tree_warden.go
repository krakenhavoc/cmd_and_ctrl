package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kin-Tree Warden — Creature — Human Warrior {G}, 1/1 (Khans of Tarkir,
// #5554):
//
//	"{2}: Regenerate this creature.
//	 Morph {G} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its morph cost.)"
//
// The shared regenerateThisRow (Soldevi Steam Beast's {2}: a
// regeneration shield on this creature, CR 701.19a). Morph is the
// keyword's declaration (morph.go, CR 702.37): cast face down for {3} it
// is a nameless 2/2 with no text (CR 708.2a), and {G} turns it face up
// as the same object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "1ea9517a-c2ed-4aec-8fd3-e603b0e7dec9",
		Name:             "Kin-Tree Warden",
		Completeness:     effects.CompletenessFull,
		Activated:        []effects.ActivatedAbility{effects.RegenerateThisRow("{2}")},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{G}")},
	})
}
