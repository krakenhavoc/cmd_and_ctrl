package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elfhame Sanctuary — Enchantment {1}{G} (Invasion, S82 #5616):
//
//	"At the beginning of your upkeep, you may search your library for a basic
//	 land card, reveal that card, put it into your hand, then shuffle. If you
//	 do, you skip your draw step this turn."
//
// "If you do" is searching: choosing to search is what skips the draw step,
// whether or not a land is found. The skip is Ivory Gargoyle's one-shot record
// (Game.SkipNextStepForEffect, #6144): the draw step is the next step after the
// upkeep, so it is this turn's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f98f270-f6ff-4029-bcdd-3e7f6483eb5d",
		Name:         "Elfhame Sanctuary",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Elfhame Sanctuary — search for a basic land card, put it into your hand, then shuffle; skip your draw step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.SkipNextStepForEffect(item.Controller, game.StepDraw, item.SourceCardID, "Elfhame Sanctuary — you skip your draw step this turn")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Elfhame Sanctuary — search for a basic land card",
						Source:    item.SourceCardID,
					}.Apply(ctx)
				}),
				"Elfhame Sanctuary — search for a basic land card (you skip your draw step)?"),
		},
	})
}
