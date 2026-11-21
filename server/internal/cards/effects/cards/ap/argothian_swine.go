package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argothian Swine — Creature — Boar {3}{G}, 3/3 (Urza's Saga, #5624):
//
//	"Trample"
//
// Registered so the creature resolves and enters on its own rather than
// through the manual sandbox path. Trample is the printed keyword the
// combat-damage assignment reads (CR 702.19a); the cost, the type and the
// 3/3 ride the card from Scryfall.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "80e08ee2-4272-4783-9a0d-90079600b594",
		Name:            "Argothian Swine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
