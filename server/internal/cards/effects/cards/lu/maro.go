package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maro — Creature — Elemental {2}{G}{G}, */* (Secret Lair Drop):
//
//	"Maro's power and toughness are each equal to the number of cards in your
//	 hand."
//
// A characteristic-defining ability (CR 604.3, layer 7a) over the controller's
// hand size, counted live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef32258a-d567-4fbc-ab85-8c00080f5b36",
		Name:         "Maro",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB021CardsInHandPT("Maro's power and toughness are each equal to the number of cards in your hand."),
		},
	})
}
