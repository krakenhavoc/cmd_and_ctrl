package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raugrin Crystal — Artifact {3} (Ikoria: Lair of Behemoths, #5512):
//
//	"{T}: Add {U}, {R}, or {W}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A mana rock for its wedge, declared like the Triomes' mana ability,
// and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae4a8470-4517-42c7-a7f6-ef7635e10fe5",
		Name:         "Raugrin Crystal",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|R|W}",
			Label:    "{T}: Add {U}, {R}, or {W}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
