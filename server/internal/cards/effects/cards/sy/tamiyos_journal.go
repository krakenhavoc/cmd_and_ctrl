package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tamiyo's Journal — Legendary Artifact — Book {5} (EDHREC rank 3016):
//
//	"At the beginning of your upkeep, investigate. (Create a Clue
//	 token. It's an artifact with "{2}, Sacrifice this token: Draw a
//	 card.")
//	 {T}, Sacrifice three Clues: Search your library for a card, put
//	 that card into your hand, then shuffle."
//
// A Clue engine whose payoff is a tutor. Investigate is the shared
// Clue template on an upkeep trigger. The tutor's cost is a tap plus
// a sacrifice clause with a count of three (#747, SacrificeN) matching
// the Clue subtype, so a nontoken Clue counts. The search does not
// reveal, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20591a83-7138-4876-b088-035bd3788be3",
		Name:         "Tamiyo's Journal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Tamiyo's Journal — investigate", effects.Do(effects.Investigate{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice three Clues: Search your library for a card, put that card into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(3, "three Clues", effects.HasSubtype("Clue"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:  item.Controller,
					Dest:    game.ZoneHand,
					Limit:   1,
					Shuffle: true,
					Source:  item.SourceCardID,
					Reason:  "Tamiyo's Journal — a card, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
