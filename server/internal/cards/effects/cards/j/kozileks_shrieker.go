package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek's Shrieker — Creature — Eldrazi Drone {2}{B}, 3/2 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {C}: This creature gets +1/+0 and gains menace until end of turn.
//	 (It can't be blocked except by two or more creatures. {C}
//	 represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The pump and the menace are
// Flame-Chain Mauler's row. {C} is colorless mana only, never generic
// (CR 107.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2b5683b-f8ea-4af6-b847-953ab4a267c4",
		Name:            "Kozilek's Shrieker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{C}: This creature gets +1/+0 and gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{C}"),
			Effect:  effects.ThisCreatureUntilEOT("Kozilek's Shrieker — +1/+0 and menace", 1, 0, "menace"),
		}},
	})
}
