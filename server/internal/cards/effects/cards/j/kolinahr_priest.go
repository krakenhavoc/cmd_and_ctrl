package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kolinahr Priest — Creature — Vulcan Monk {2}{U}, 3/2 (Star Trek
// Commander, #4974):
//
//	"Flash"
//
// The printed keyword: the cast gate lets it be cast any time its
// controller could cast an instant (CR 702.8a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b03c6720-b899-4eba-b28d-54ded03e086d",
		Name:            "Kolinahr Priest",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
