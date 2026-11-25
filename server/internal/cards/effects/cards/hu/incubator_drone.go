package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incubator Drone — Creature — Eldrazi Drone {3}{U}, 2/3 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, create a 1/1 colorless Eldrazi Scion creature
//	 token. It has "Sacrifice this token: Add {C}.""
//
// Scion Summoner's trigger, on the catalog's Eldrazi Scion.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "088d63ec-bd5e-4e80-acf9-2af270a9b395",
		Name:            "Incubator Drone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Incubator Drone — create a 1/1 colorless Eldrazi Scion",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1})),
				game.Purpose{Tokens: 1}),
		},
	})
}
