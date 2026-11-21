package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lionheart Maverick — Creature — Human Knight {W}, 1/1 (Guildpact,
// #5599):
//
//	"Vigilance
//	 {4}{W}: This creature gets +1/+2 until end of turn."
//
// Hematite Golem's self-pump (thisGetsUntilEndOfTurn) at {4}{W} for
// +1/+2, on a vigilant body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5058b9f1-9941-495d-9eed-554ae2f6d500",
		Name:            "Lionheart Maverick",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: This creature gets +1/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 2}},
			Cost:    effects.ManaCost("{4}{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 2, "Lionheart Maverick — +1/+2 until end of turn"),
		}},
	})
}
