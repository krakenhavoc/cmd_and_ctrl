package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cavern Crawler — Creature — Insect {2}{R}, 0/3 (Mercadian Masques, S82 OLD,
// #5619):
//
//	"Mountainwalk
//	 {R}: This creature gets +1/-1 until end of turn."
//
// Mountainwalk is printed data. The ability can be activated as often as there
// is red mana, each bonus shrinking toughness: at toughness 0 the Crawler dies as
// a state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7cfb120-1344-4b4e-9c6f-f969ba9c4a18",
		Name:            "Cavern Crawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Cavern Crawler — +1/-1 until end of turn"),
		}},
	})
}
