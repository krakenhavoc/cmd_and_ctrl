package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonwing Moth — Creature — Insect {1}{W}{W}, 2/1 (Saviors of Kamigawa,
// #5601):
//
//	"Flying
//	 {W}: This creature gets +0/+1 until end of turn."
//
// Basalt Gargoyle's toughness pump on white mana: a layer 7c boost on this
// creature that ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6501b3c7-fd21-4668-b647-db12f0cc23f7",
		Name:            "Moonwing Moth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(0, 1, "Moonwing Moth — +0/+1"),
		}},
	})
}
