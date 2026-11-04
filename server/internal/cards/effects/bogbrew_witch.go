package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bogbrew Witch — Creature — Human Wizard {3}{B}, 1/3:
//
//	"{2}, {T}: Search your library for a card named Festering Newt or
//	 Bubbling Cauldron, put it onto the battlefield tapped, then
//	 shuffle."
//
// A search for a stated quality, so the searcher may find nothing (CR
// 701.23b) and the library is shuffled either way. The found card
// enters tapped; it is a permanent card by its name, so it always has
// somewhere to go. The Witch's {T} means summoning sickness applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad138184-ab2d-407a-9f9c-94b0675be19d",
		Name:         "Bogbrew Witch",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Search your library for a card named Festering Newt or Bubbling Cauldron, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return game.HasName(c, "Festering Newt") || game.HasName(c, "Bubbling Cauldron")
					},
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Source:        item.SourceCardID,
					Reason:        "Bogbrew Witch — search for a card named Festering Newt or Bubbling Cauldron",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
