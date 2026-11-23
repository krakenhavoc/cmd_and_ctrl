package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky-Eel School — Creature — Fish {3}{U}{U}, 3/3 (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, draw a card, then discard a card."
//
// The loot is the shared draw-then-discard body (b16DrawThenDiscard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c90ea9d5-82f7-4b92-8cb1-10d0a7378e8c",
		Name:            "Sky-Eel School",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sky-Eel School — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) }),
		},
	})
}
