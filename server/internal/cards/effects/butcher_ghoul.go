package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Butcher Ghoul — Creature — Zombie {1}{B}, 1/1 (Innistrad Remastered, #5948):
//
//	"Undying (When this creature dies, if it had no +1/+1 counters on it, return
//	 it to the battlefield under its owner's control with a +1/+1 counter on it.)"
//
// Undying is the printed keyword (PrintedKeywords, game/undying_persist.go): the
// engine reads it off the card, so nothing else is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28caeed4-c175-469f-8d8e-f45bc0a51a1c",
		Name:            "Butcher Ghoul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
	})
}
