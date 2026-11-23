package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sivriss, Nightmare Speaker — {3}{B} Legendary Creature — Snake Cleric
// Warlock 3/3
//
//	"{T}, Sacrifice another creature or an artifact: For each opponent, you
//	 mill a card, then return that card from your graveyard to your hand
//	 unless that player pays 3 life.
//	 Choose a Background"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed15a3c4-110e-4f08-bf09-fabbf5b4eeb9",
		Name:         "Sivriss, Nightmare Speaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature or an artifact: For each opponent, you mill a card, then return that card from your graveyard to your hand unless that player pays 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.TapCost(),
				effects.SacrificeAnotherN(1, "another creature or an artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			Effect: effects.GapB213SivrissMillThenReturnUnlessPaid,
		}},
	})
}
