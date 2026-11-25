package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruination Guide — Creature — Eldrazi Drone {2}{U}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)
//	 Other colorless creatures you control get +1/+0."
//
// Kozilek, the Broken Reality's anthem over the other colorless creatures
// you control, at +1/+0; ingest is OldB305Ingest.
//
// No simplification.
func init() {
	const name = "Ruination Guide"
	effects.Register(effects.Spec{
		OracleID:        "3394ce74-c352-4ae1-a7bb-f6ef8ac1c810",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered:       []game.TriggeredAbility{effects.OldB305Ingest(name)},
		Static:          []game.StaticAbility{effects.OldB305OtherColorlessAnthem(1, 0, "Other colorless creatures you control get +1/+0.")},
	})
}
