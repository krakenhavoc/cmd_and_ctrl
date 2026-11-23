package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whirlwind Adept — Creature — Djinn Monk {4}{U}, 4/2 (Khans of Tarkir,
// #5554):
//
//	"Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Hexproof is an engine keyword (CR 702.11b): an opponent's spell or
// ability can't target it. Prowess is derived by the engine from the
// keyword (game/prowess.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c7e261f-45d5-4c86-8cba-9b6edce8ef49",
		Name:            "Whirlwind Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof", "prowess"},
	})
}
