package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liliana's Shade — Creature — Shade {2}{B}{B}, 1/1 (Magic 2013):
//
//	"When this creature enters, you may search your library for a Swamp card,
//	 reveal it, put it into your hand, then shuffle.
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Moon-Blessed Cleric's optional search (the prompt carries "Don't search",
// which skips the search and the shuffle outright; finding nothing is a search
// that came up empty and still shuffles, CR 701.23) for a Swamp card, which is
// a land with the subtype, and Ridgeline Rager's firebreathing at {B} with both
// halves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44f87b6e-0cc8-4088-8a94-8272b6b26869",
		Name:         "Liliana's Shade",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Liliana's Shade — search for a Swamp card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Swamp") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Liliana's Shade — a Swamp card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Liliana's Shade — +1/+1"),
		}},
	})
}
