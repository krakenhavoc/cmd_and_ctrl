package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prophet of Distortion — Creature — Eldrazi Drone {U}, 1/2 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {3}{C}: Draw a card. ({C} represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). {C} is colorless mana only,
// never generic (CR 107.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6bc82093-0246-4c60-bf30-0b605bef01cd",
		Name:            "Prophet of Distortion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{C}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.ManaCost("{3}{C}"),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
