package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Kinsmith — Artifact Creature — Myr {4}, 3/1:
//
//	"When this creature enters, you may search your library for a Myr
//	 card, reveal it, put it into your hand, then shuffle."
//
// A search naming a quality may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97bb382b-16b2-44db-a2bd-de09ca90bfcc",
		Name:         "Myr Kinsmith",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Myr Kinsmith — search for a Myr card and put it into your hand", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Myr") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Myr Kinsmith — a Myr card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
