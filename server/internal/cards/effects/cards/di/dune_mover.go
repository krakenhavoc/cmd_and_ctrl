package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dune Mover — Artifact Creature — Phyrexian Golem {2}, 2/1:
//
//	"Toxic 1
//	 When this creature enters, you may search your library for a basic
//	 land card, reveal it, then shuffle and put that card on top."
//
// Compass Gnome's search: the card never leaves the library and is placed
// after the shuffle. Toxic is the engine's damage tail.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "38d2176c-1a1e-4030-9c63-e28ba4a69b6d",
		Name:            "Dune Mover",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tutors: 1},
		PrintedKeywords: []string{"toxic 1"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dune Mover — search for a basic land card and put it on top", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Dune Mover — a basic land card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
