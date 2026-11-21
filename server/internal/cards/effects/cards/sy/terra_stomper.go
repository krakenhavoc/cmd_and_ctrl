package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terra Stomper — Creature — Beast {3}{G}{G}{G}, 8/8 (Zendikar, #5583):
//
//	"This spell can't be countered.
//	 Trample (This creature can deal excess combat damage to the player
//	 or planeswalker it's attacking.)"
//
// The can't-be-countered clause is the Spec flag the counterspell path
// reads (CR 101.2); trample is the printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d688baac-0574-4d88-b7eb-a7384c298ef5",
		Name:            "Terra Stomper",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample"},
	})
}
