package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brood Monitor — {4}{G}{G} Creature — Eldrazi Drone 3/3 (Jumpstart 2022, #5948):
//
//	"Devoid
//	 When this creature enters, create three 1/1 colorless Eldrazi Scion
//	 creature tokens. They have "Sacrifice this token: Add {C}.""
//
// Scion Summoner's token, three of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a08bb4b-46df-4b99-8f86-670d07e900ba",
		Name:            "Brood Monitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Brood Monitor — create three 1/1 colorless Eldrazi Scion creature tokens",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 3})),
				game.Purpose{Tokens: 3}),
		},
	})
}
