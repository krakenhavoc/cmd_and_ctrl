package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mist Intruder — Creature — Eldrazi Drone {1}{U}, 1/2 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)"
//
// Ingest is OldB305Ingest.
//
// No simplification.
func init() {
	const name = "Mist Intruder"
	effects.Register(effects.Spec{
		OracleID:        "e4e9b48b-92f9-46ac-b55a-573bc328b11a",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered:       []game.TriggeredAbility{effects.OldB305Ingest(name)},
	})
}
