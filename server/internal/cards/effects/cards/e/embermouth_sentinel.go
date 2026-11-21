package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embermouth Sentinel — Artifact Creature — Chimera {2}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, you may search your library for a basic
//	 land card, reveal it, then shuffle and put that card on top. If you
//	 control a Dragon, put that card onto the battlefield tapped instead."
//
// Compass Gnome's search-and-put-on-top, with the Dragon check made as the
// ability resolves (layers applied, so a changeling or a Dragon an effect
// made counts): with a Dragon the same search puts the land onto the
// battlefield tapped instead of on top, the library shuffled either way.
// "You may" lets the searcher decline both the card and the shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5208c2c-b839-42f6-8717-ee4bbaf1385b",
		Name:         "Embermouth Sentinel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Embermouth Sentinel — you may search for a basic land card and put it on top, or onto the battlefield tapped if you control a Dragon",
				func(g *game.Game, item *game.StackItem) error {
					search := effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Limit:     1,
						Optional:  true,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Embermouth Sentinel — a basic land card",
					}
					if effects.StdB9YouControlA(g, item.Controller, effects.HasSubtype("Dragon")) {
						search.Dest = game.ZoneBattlefield
						search.TappedOnEntry = true
					} else {
						search.Dest = game.ZoneLibrary
						search.ToTop = true
					}
					return search.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
