package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stoic Farmer — Creature — Dwarf Peasant {3}{W}, 3/3 (Kaldheim Commander,
// #5506):
//
//	"When this creature enters, search your library for a basic Plains card
//	 and reveal it. If an opponent controls more lands than you, put it onto
//	 the battlefield tapped. Otherwise, put it into your hand. Then shuffle.
//	 Foretell {1}{W}"
//
// The land count is read as the trigger resolves, before the search (nothing
// in it changes a land count), and "an opponent" is any one of them. A basic
// Plains is read by supertype, so a Snow-Covered Plains qualifies. Foretell
// is the engine's special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "42820e09-f817-4485-a96a-ac2b0b798269",
		Name:           "Stoic Farmer",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stoic Farmer — search for a basic Plains; onto the battlefield tapped if an opponent controls more lands, otherwise to your hand",
				func(g *game.Game, item *game.StackItem) error {
					dest, tapped := game.ZoneHand, false
					if effects.AnOpponentHasMoreLands(g, item.Controller) {
						dest, tapped = game.ZoneBattlefield, true
					}
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLandWithSubtype("plains"),
						Dest:          dest,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: tapped,
						Reason:        "Stoic Farmer — a basic Plains card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
