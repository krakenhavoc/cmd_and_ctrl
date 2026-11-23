package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sporemound — Creature — Fungus {3}{G}{G}, 3/3 (EDHREC rank 3591):
//
//	"Landfall — Whenever a land you control enters, create a 1/1
//	 green Saproling creature token."
//
// The landfall Saproling engine. Fires once per land, however it
// entered — played, fetched, or returned from a graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1be56a3d-a6c0-4b65-ae71-3d90ceefc6c0",
		Name:         "Sporemound",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Sporemound — create a 1/1 green Saproling", game.LandfallPayoff{Tokens: 1}, effects.B34CreateTokens(effects.B11GreenSaprolingToken, 1)),
		},
	})
}
