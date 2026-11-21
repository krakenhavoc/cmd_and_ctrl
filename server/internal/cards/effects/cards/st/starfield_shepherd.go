package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starfield Shepherd — Creature — Angel {3}{W}{W}, 3/2 (Edge of
// Eternities, #4461):
//
//	"Flying
//	 When this creature enters, search your library for a basic Plains
//	 card or a creature card with mana value 1 or less, reveal it, put it
//	 into your hand, then shuffle.
//	 Warp {1}{W}"
//
// The search is Gloryheath Lynx's "basic Plains card" (the Basic
// supertype, so a Snow-Covered Plains is one and a nonbasic Plains is
// not) or a creature card whose mana value is 1 or less (CR 202.3). It
// is not a "may": the search is mandatory, though it can find nothing.
// Warp is the shared alternative cost (CR 702.185), so a warped Shepherd
// still searches as it enters and is exiled at the next end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "42fa99ab-d35f-4439-ab00-fe99bad626ac",
		Name:             "Starfield Shepherd",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Starfield Shepherd — search for a basic Plains card or a creature card with mana value 1 or less",
				func(g *game.Game, item *game.StackItem) error {
					plains := effects.IsBasicLandWithSubtype("plains")
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return plains(c) || (c.IsCreature() && c.ManaValue() <= 1)
						},
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  "Starfield Shepherd — a basic Plains card or a creature card with mana value 1 or less",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
