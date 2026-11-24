package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of Uthuun — Creature — Sphinx {5}{U}{U}, 5/6 (Commander Legends,
// #5507):
//
//	"Flying
//	 When this creature enters, reveal the top five cards of your library.
//	 An opponent separates those cards into two piles. Put one pile into
//	 your hand and the other into your graveyard."
//
// Fact or Fiction's reveal, split and pile pick (Sphinx of Clear Skies is
// the creature precedent) on an enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b4f268d-34ae-416d-a412-66f4fe37659c",
		Name:            "Sphinx of Uthuun",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sphinx of Uthuun — reveal the top five cards of your library; an opponent separates them into two piles; put one pile into your hand and the other into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return effects.OldB081FactOrFictionReveal(effects.NewContext(g, item), 5)
				}),
		},
	})
}
