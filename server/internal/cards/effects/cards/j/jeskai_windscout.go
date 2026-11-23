package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeskai Windscout — Creature — Bird Scout {2}{U}, 2/1 (Khans of Tarkir,
// #5554):
//
//	"Flying
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Flying is printed data (CR 702.9b). Prowess is derived by the engine
// from the keyword (game/prowess.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c35f4b80-90b6-467a-a666-1cc3437d49ab",
		Name:            "Jeskai Windscout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
	})
}
