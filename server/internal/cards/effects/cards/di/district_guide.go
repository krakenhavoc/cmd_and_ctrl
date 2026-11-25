package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// District Guide — Creature — Elf Scout {2}{G}, 2/2 (Guilds of Ravnica):
//
//	"When this creature enters, you may search your library for a basic land
//	 card or Gate card, reveal it, put it into your hand, then shuffle."
//
// Farfinder's optional fetch with Open the Gates' predicate: a basic land
// card, or any card with the Gate subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1eb4c001-1115-4b99-bbbf-703eb551a381",
		Name:         "District Guide",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("District Guide — you may search your library for a basic land card or Gate card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return effects.IsBasicLand(c) || c.HasSubtype("Gate") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "District Guide — a basic land card or Gate card, to your hand",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
