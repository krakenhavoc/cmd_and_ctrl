package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Birthing Hulk — {6}{G} Creature — Eldrazi Drone 5/4 (Oath of the Gatewatch,
// S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, create two 1/1 colorless Eldrazi Scion
//	 creature tokens. They have "Sacrifice this token: Add {C}."
//	 {1}{C}: Regenerate this creature."
//
// Brood Monitor's Scion tokens, two of them, and Lorwyn's regenerate row with
// a {C} in its cost, which only colorless mana pays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "33899675-0de2-4adb-bf3f-55e18b493015",
		Name:            "Birthing Hulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Birthing Hulk — create two 1/1 colorless Eldrazi Scion creature tokens",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 2})),
				game.Purpose{Tokens: 2}),
		},
		Activated: []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{1}{C}")},
	})
}
