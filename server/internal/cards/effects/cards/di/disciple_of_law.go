package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disciple of Law — Creature — Human Cleric {1}{W}, 1/2 (Urza's Saga):
//
//	"Protection from red
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Disciple of Grace with protection from red.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a4689767-46c3-42ab-af27-5877e226bc05",
		Name:            "Disciple of Law",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
