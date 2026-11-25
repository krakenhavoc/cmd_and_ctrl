package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Wanderer — Creature — Elf Druid {3}{G}, 3/2 (Kaladesh, #5539):
//
//	"When this creature enters, you may search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Omen of the Hunt's enters trigger (effects.B07SearchBasicOntoBattlefield, tapped
// and optional): the "you may" is the search prompt's decline, which skips
// the shuffle too. Nothing is revealed: the card doesn't say to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff16f889-9203-4874-919b-42e540cadcbc",
		Name:         "Wild Wanderer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wild Wanderer — you may search for a basic land, tapped", func(g *game.Game, item *game.StackItem) error {
				return effects.B07SearchBasicOntoBattlefield(g, item, item.Controller, true, true, "Wild Wanderer — a basic land card, onto the battlefield tapped")
			}),
		},
	})
}
