package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ayumi, the Last Visitor — Legendary Creature — Spirit {3}{G}{G}, 7/3
// (Saviors of Kamigawa, #5601):
//
//	"Legendary landwalk (This creature can't be blocked as long as
//	 defending player controls a legendary land.)"
//
// Legendary landwalk is an engine keyword (CR 702.14c, game/landwalk.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12acd466-deb9-4e68-b75b-aed86680b917",
		Name:            "Ayumi, the Last Visitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"legendary landwalk"},
	})
}
