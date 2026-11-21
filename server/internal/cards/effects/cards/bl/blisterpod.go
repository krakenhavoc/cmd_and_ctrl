package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blisterpod — {G} Creature — Eldrazi Drone 1/1 (Jumpstart 2022, #5948):
//
//	"Devoid
//	 When this creature dies, create a 1/1 colorless Eldrazi Scion creature
//	 token. It has "Sacrifice this token: Add {C}.""
//
// Scion Summoner's token (the printed Scion with its mana ability) made by
// a dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a62b226-d317-4212-9781-9c1d72056f11",
		Name:            "Blisterpod",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Blisterpod — create a 1/1 colorless Eldrazi Scion",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1})),
				game.Purpose{Tokens: 1}),
		},
	})
}
