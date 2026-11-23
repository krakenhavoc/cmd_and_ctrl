package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telepathy — Enchantment {U}:
//
//	"Your opponents play with their hands revealed."
//
// #3953, ADR 0199: every opponent's hand is shown to every player while
// Telepathy is on the battlefield (game.HandRevealedLocked), the cards
// they draw included. Its controller's own hand stays hidden.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d2e647ac-717f-46e6-baa8-6a793e7bdc32",
		Name:          "Telepathy",
		Completeness:  effects.CompletenessFull,
		HandsRevealed: game.HandRevealOpponents,
	})
}
