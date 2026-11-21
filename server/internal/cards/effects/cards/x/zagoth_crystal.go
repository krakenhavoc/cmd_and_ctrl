package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zagoth Crystal — Artifact {3} (Ikoria: Lair of Behemoths, #5512):
//
//	"{T}: Add {B}, {G}, or {U}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A mana rock for its wedge, declared like the Triomes' mana ability,
// and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87edb566-8929-4961-9c03-d3657a8c6feb",
		Name:         "Zagoth Crystal",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|G|U}",
			Label:    "{T}: Add {B}, {G}, or {U}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
