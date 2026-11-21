package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slaughter Drone — Creature — Eldrazi Drone {1}{B}, 2/2 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {C}: This creature gains deathtouch until end of turn. (Any amount
//	 of damage it deals to a creature is enough to destroy it. {C}
//	 represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The deathtouch is a layer-6
// grant to this creature until cleanup. {C} is colorless mana only,
// never generic (CR 107.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "047e09c0-cbd9-45ba-9729-bf35ebe24e69",
		Name:            "Slaughter Drone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{C}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{C}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Slaughter Drone — deathtouch until end of turn"),
		}},
	})
}
