package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Citanul Flute — Artifact {5} (Urza's Saga):
//
//	"{X}, {T}: Search your library for a creature card with mana value X
//	 or less, reveal it, put it into your hand, then shuffle."
//
// Celestial Reunion's search as an activated ability: X is the value the
// ability was activated for, and the found card is revealed and goes to
// the hand. Finding nothing is allowed; the library is shuffled either
// way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b8183bb-20c4-4d45-b944-508a6a874bb3",
		Name:         "Citanul Flute",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: Search your library for a creature card with mana value X or less, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= x },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Citanul Flute — a creature card with mana value X or less",
				}.Apply(ctx)
			},
		}},
	})
}
