package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran Outrider — Creature — Human Soldier {1}{W}, 2/2 (Coldsnap, S82 OLD b560):
//
//	"{W}: This creature gets +0/+1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b8559a7-e11f-430d-8a7a-785a7005c4c4",
		Name:         "Kjeldoran Outrider",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{W}", 0, 1, "Kjeldoran Outrider")},
	})
}
