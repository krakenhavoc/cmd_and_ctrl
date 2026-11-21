package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scouting Hawk — Creature — Bird {2}{W}, 1/1 (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Flying
//	 Keen Sight — When this creature enters, if an opponent controls more
//	 lands than you, search your library for a basic Plains card, put it
//	 onto the battlefield tapped, then shuffle."
//
// An intervening "if" (CR 603.4): it is read when the Hawk enters and again
// as the trigger resolves. The search is not "may": the library is searched
// whether or not a Plains is there, then shuffled; a basic Plains is a
// basic land card with the Plains subtype (Keeper of the Accord's search).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f96c48f-aceb-4554-8b06-cca162151a64",
		Name:            "Scouting Hawk",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Lands: 1, Tutors: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.IfAnOpponentControlsMoreLandsThanYou(effects.WhenThisEnters("Scouting Hawk — Keen Sight: search for a basic Plains card onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return effects.IsBasicLand(c) && c.HasSubtypeWord("Plains")
						},
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Scouting Hawk — a basic Plains card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
