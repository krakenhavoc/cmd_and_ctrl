package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furnace Host Charger — Creature — Phyrexian Giant {5}{R}, 5/5 (March of
// the Machine):
//
//	"Haste
//	 Mountaincycling {2} ({2}, Discard this card: Search your library for
//	 a Mountain card, reveal it, put it into your hand, then shuffle.)"
//
// Mountaincycling is the shared Typecycling constructor (CR 702.29e), as on
// Chartooth Cougar.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fea60c45-cc36-4dd1-8475-23a9361e370d",
		Name:            "Furnace Host Charger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}
