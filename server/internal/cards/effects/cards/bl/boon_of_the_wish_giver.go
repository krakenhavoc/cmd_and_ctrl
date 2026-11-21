package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boon of the Wish-Giver — Sorcery {4}{U}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Draw four cards.
//	 Cycling {1} ({1}, Discard this card: Draw a card.)"
//
// Tidings' draw, and cycling from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1720d38a-c3c1-4294-85c1-a91fc7179d6f",
		Name:         "Boon of the Wish-Giver",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve:    effects.ResolveSteps(effects.DrawCards{N: 4}),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}")},
	})
}
