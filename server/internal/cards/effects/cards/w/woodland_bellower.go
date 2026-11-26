package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Woodland Bellower — Creature — Beast {4}{G}{G}, 6/5 (Magic Origins):
//
//	"When this creature enters, you may search your library for a
//	 nonlegendary green creature card with mana value 3 or less, put it
//	 onto the battlefield, then shuffle."
//
// An enters trigger with the search chooser. The filter reads the card's
// effective colours, supertypes and mana value; "you may" is the prompt's
// decline. The found card is not revealed (nothing prints it) and enters
// untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5859fd59-8732-414a-8932-007980af495d",
		Name:         "Woodland Bellower",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Woodland Bellower — search for a nonlegendary green creature card with mana value 3 or less", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && effects.B34IsGreen(c) && !effects.IsLegendary(&c) && c.ManaValue() <= 3
					},
					Dest:     game.ZoneBattlefield,
					Limit:    1,
					Shuffle:  true,
					Optional: true,
					Reason:   "Woodland Bellower — a nonlegendary green creature card with mana value 3 or less",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
