package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valiant Rescuer — Creature — Human Soldier {1}{W}, 3/1 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Whenever you cycle another card for the first time each turn, create a
//	 1/1 white Human Soldier creature token.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The trigger reads the turn's event log: it fires on a cycle of a card
// other than the Rescuer when no other card was cycled by that player
// earlier this turn. The Rescuer's own cycling happens in the hand, where
// it watches nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9de9f0b-bf07-4171-b54d-69033424a212",
		Name:         "Valiant Rescuer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCycle, effects.S82OldB119CycledAnotherCardForTheFirstTimeThisTurn,
				"Valiant Rescuer — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
