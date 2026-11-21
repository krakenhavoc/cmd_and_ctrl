package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider-Bot — Artifact Creature — Spider Robot Scout {2}, 2/1 (Marvel's
// Spider-Man, #4460):
//
//	"Reach
//	 When this creature enters, you may search your library for a basic land
//	 card, reveal it, then shuffle and put that card on top."
//
// Moon-Blessed Cleric's search: SearchLibrary with ToTop finds, reveals,
// shuffles and then puts the card on top, and Optional is the printed "you may"
// (a declined search does not look, shuffle or trigger searchers).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4bb33b2-adde-4176-9274-7598b559aa6a",
		Name:            "Spider-Bot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spider-Bot — search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Spider-Bot — a basic land card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
