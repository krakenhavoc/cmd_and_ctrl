package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeskai Student — Creature — Human Monk {1}{W}, 1/3 (Khans of Tarkir,
// #5554):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Prowess is derived by the engine from the keyword (game/prowess.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8ebfc8b-3cee-491b-a568-8a8dcb22541d",
		Name:            "Jeskai Student",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
