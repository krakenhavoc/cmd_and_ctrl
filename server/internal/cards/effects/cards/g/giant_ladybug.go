package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Ladybug — {2}{G} Creature — Insect 4/1 (Jumpstart 2022, #5948):
//
//	"Reach
//	 When this creature enters, you may search your library for a basic land
//	 card, reveal it, then shuffle and put that card on top."
//
// The one-mana tutors' shuffle-then-top search (ToTop), basic lands only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "abfe31f1-7c6d-4a7a-9f44-7f3e4166d9f1",
		Name:            "Giant Ladybug",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Giant Ladybug — you may search your library for a basic land card, reveal it, then shuffle and put that card on top",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneLibrary,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						ToTop:     true,
						Optional:  true,
						Source:    item.SourceCardID,
						Reason:    "Giant Ladybug — a basic land card, to the top of your library",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
