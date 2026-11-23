package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6b8183bb-20c4-4d45-b944-508a6a874bb3",
		Name:         "Citanul Flute",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Search your library for a creature card with mana value X or less, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.X()
				return effects.SearchLibrary{
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
