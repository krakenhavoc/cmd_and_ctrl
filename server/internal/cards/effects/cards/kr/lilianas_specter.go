package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liliana's Specter — Creature — Specter {1}{B}{B}, 2/1 (Magic 2011,
// #5579):
//
//	"Flying
//	 When this creature enters, each opponent discards a card."
//
// Flying is the printed keyword. The enters trigger is Burglar Rat's:
// every opponent picks and discards one card as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1a7e0e55-41c3-4fd8-b718-55b59b5cab90",
		Name:            "Liliana's Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Liliana's Specter — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
