package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien's Arkbow — Legendary Artifact {1}{G} (War of the Spark, #5520):
//
//	"{X}, {T}, Discard a card: Look at the top X cards of your library. You
//	 may put a creature card with mana value X or less from among them onto
//	 the battlefield. Put the rest on the bottom of your library in a random
//	 order."
//
// Muzzio's look and put (PutFromLibraryOntoBattlefield) with the creature's
// mana value bounded by the X paid; the discard is a cost, the rest go to the
// bottom in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efe3204d-2013-47a1-ad68-adfbb2d0be8f",
		Name:         "Vivien's Arkbow",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}, Discard a card: Look at the top X cards of your library. You may put a creature card with mana value X or less from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.DiscardACard()),
			Effect:  effects.S82OldB172ArkbowDig,
		}},
	})
}
