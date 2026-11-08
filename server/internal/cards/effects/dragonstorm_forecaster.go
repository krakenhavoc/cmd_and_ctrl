package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonstorm Forecaster — Creature — Human Scout {U}, 0/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"{2}, {T}: Search your library for a card named Dragonstorm Globe or
//	 Boulderborn Dragon, reveal it, put it into your hand, then shuffle."
//
// Magitek Infantry's name search, to hand and revealed, with two names.
// The search may find nothing (CR 701.23b), and the library is shuffled
// either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02de8fa3-ac8d-47e5-b137-d5a0b7d8e951",
		Name:         "Dragonstorm Forecaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Search your library for a card named Dragonstorm Globe or Boulderborn Dragon, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{2}", Tap: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return game.HasName(c, "Dragonstorm Globe") || game.HasName(c, "Boulderborn Dragon")
					},
					Dest:    game.ZoneHand,
					Limit:   1,
					Reveal:  true,
					Shuffle: true,
					Source:  item.SourceCardID,
					Reason:  "Dragonstorm Forecaster — search for a card named Dragonstorm Globe or Boulderborn Dragon",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
