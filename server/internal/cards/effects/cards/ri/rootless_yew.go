package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootless Yew — Creature — Treefolk {3}{G}{G}, 5/4 (Kaldheim):
//
//	"When this creature dies, search your library for a creature card with
//	 power or toughness 6 or greater, reveal it, put it into your hand, then
//	 shuffle."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b1dc298-403d-484d-90dd-e1c9b73e0188",
		Name:         "Rootless Yew",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Rootless Yew — search your library for a creature card with power or toughness 6 or greater, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsCreature() && (c.Power >= 6 || c.Toughness >= 6) },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Rootless Yew — a creature card with power or toughness 6 or greater",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
