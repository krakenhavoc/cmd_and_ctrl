package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Culling Drone — Creature — Eldrazi Drone {1}{B}, 2/2 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)"
//
// Ingest is OldB307Ingest.
//
// No simplification.
func init() {
	const name = "Culling Drone"
	effects.Register(effects.Spec{
		OracleID:        "bda0ca68-78a6-4b32-842f-0bee9a9c84ef",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered:       []game.TriggeredAbility{effects.OldB307Ingest(name)},
	})
}
