package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honored Knight-Captain — Creature — Human Advisor Knight {1}{W}, 1/1
// (Edge of Eternities, #4461):
//
//	"When this creature enters, create a 1/1 white Human Soldier creature
//	 token.
//	 {4}{W}{W}, Sacrifice this creature: Search your library for an
//	 Equipment card, put it onto the battlefield, then shuffle."
//
// The entry trigger is Knight Luminary's token. The activated ability is
// Sakura-Tribe Elder's sacrifice-and-search with an Equipment card put
// onto the battlefield untapped, unattached. Finding nothing is allowed
// (CR 701.23b), and the library is shuffled either way. It can be
// activated at instant speed and answers nothing, so it declares
// AnswerValue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69741d3b-1259-43f5-897f-6f4c3ca0844b",
		Name:         "Honored Knight-Captain",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Honored Knight-Captain — create a 1/1 white Human Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}{W}, Sacrifice this creature: Search your library for an Equipment card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}{W}{W}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype("Equipment") },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Honored Knight-Captain — an Equipment card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
