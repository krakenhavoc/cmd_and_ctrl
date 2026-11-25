package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benthic Infiltrator — Creature — Eldrazi Drone {2}{U}, 1/4 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)
//	 This creature can't be blocked."
//
// Ingest is OldB305Ingest; "can't be blocked" is Blighted Agent's restriction
// bit, read at the block declaration (CR 509.1b).
//
// No simplification.
func init() {
	const name = "Benthic Infiltrator"
	effects.Register(effects.Spec{
		OracleID:        "b2977807-5f46-424a-a6e9-14262f8a723a",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered:       []game.TriggeredAbility{effects.OldB305Ingest(name)},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
