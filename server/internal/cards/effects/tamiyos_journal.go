package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "20591a83-7138-4876-b088-035bd3788be3",
		Name:         "Tamiyo's Journal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Tamiyo's Journal — investigate", Do(CreateToken{Template: ClueToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice three Clues: Search your library for a card, put that card into your hand, then shuffle.",
			Cost:  Plus(TapCost(), SacrificeN(3, "three Clues", HasSubtype("Clue"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(game.Card) bool { return true },
					Dest:      game.ZoneHand,
					Limit:     1,
					Shuffle:   true,
					Source:    item.SourceCardID,
					Reason:    "Tamiyo's Journal — a card, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
