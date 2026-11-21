package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wasteland Scorpion — Creature — Scorpion {2}{B}, 2/2
// (Amonkhet, #5535):
//
//	"Deathtouch
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Cycling is the shared constructor; the cycle draws through the ordinary
// activation from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0bc418c3-ebde-4e52-9284-65d7e3433f27",
		Name:            "Wasteland Scorpion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
