package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clifftop Lookout — Creature — Frog Scout {2}{G}, 1/2 (Bloomburrow,
// #4467):
//
//	"Reach
//	 When this creature enters, reveal cards from the top of your library
//	 until you reveal a land card. Put that card onto the battlefield
//	 tapped and the rest on the bottom of your library in a random
//	 order."
//
// The Regalia's sentence (RevealUntilThenPutOntoBattlefield) on an entry
// trigger. A library with no land card reveals itself entirely and every
// card goes to the bottom in a random order. Reach rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "446db5a4-21b0-44c1-b12b-d365313dd67a",
		Name:            "Clifftop Lookout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Clifftop Lookout — reveal until a land card, put it onto the battlefield tapped",
				effects.Do(effects.RevealUntilThenPutOntoBattlefield{
					Match:  game.Card.IsLand,
					Tapped: true,
					Reason: "Clifftop Lookout — revealed until a land card",
				})),
		},
	})
}
