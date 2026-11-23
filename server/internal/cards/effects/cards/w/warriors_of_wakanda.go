package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warriors of Wakanda — Creature — Human Warrior {3}{G}, 4/4 (Marvel
// Super Heroes Commander, #4455): "Trample". A keyword-only body, the
// same shape as Glamorous Grapplers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7abbbe18-a2e5-47ce-9c65-8008e5153a4d",
		Name:            "Warriors of Wakanda",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
