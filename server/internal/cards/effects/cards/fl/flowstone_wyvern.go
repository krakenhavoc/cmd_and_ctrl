package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Wyvern — Creature — Drake {3}{R}{R}, 3/3 (Tempest, S82 OLD b755,
// #5628):
//
//	"Flying
//	 {R}: This creature gets +2/-2 until end of turn."
//
// Breathstealer's self pump with a toughness loss: a Wyvern pumped past its
// toughness dies to the state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13459bf3-fcdc-4838-9669-918f97113054",
		Name:            "Flowstone Wyvern",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +2/-2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: -2}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, -2, "Flowstone Wyvern — +2/-2 until end of turn"),
		}},
	})
}
