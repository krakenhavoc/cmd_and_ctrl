package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scion Summoner — Creature — Eldrazi Drone {2}{G}, 2/2 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, create a 1/1 colorless Eldrazi Scion
//	 creature token. It has "Sacrifice this token: Add {C}." ({C}
//	 represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The Scion is the catalog's
// Eldrazi Scion (b28EldraziScionToken), whose sacrifice for {C} is a
// real mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa60f3a6-5834-4059-adea-dd17366b7885",
		Name:            "Scion Summoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Scion Summoner — create a 1/1 colorless Eldrazi Scion",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1})),
				game.Purpose{Tokens: 1}),
		},
	})
}
