package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lava Serpent — Creature — Elemental Serpent {5}{R}, 5/5 (Ikoria: Lair of Behemoths, #5512):
//
//	"Haste
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The printed keyword, and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d94d807-97d9-4cdf-a63c-d8b82c2acc6e",
		Name:            "Lava Serpent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
