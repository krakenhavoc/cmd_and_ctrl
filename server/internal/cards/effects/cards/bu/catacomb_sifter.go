package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Catacomb Sifter — Creature — Eldrazi Drone {1}{B}{G}, 2/3 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, create a 1/1 colorless Eldrazi Scion creature
//	 token. It has "Sacrifice this token: Add {C}."
//	 Whenever another creature you control dies, scry 1."
//
// Scion Summoner's token trigger, and Bereaved Survivor's "another creature you
// control dies" condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c2c5f3b-61f3-4772-9ade-f5be80719e9f",
		Name:            "Catacomb Sifter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Catacomb Sifter — create a 1/1 colorless Eldrazi Scion",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1})),
				game.Purpose{Tokens: 1}),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B39AnotherCreatureYouControlDied(ev, source, g)
			},
				"Catacomb Sifter — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
