package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burlfist Oak — Creature — Treefolk {2}{G}{G}, 2/3 (Core Set 2021, #5511):
//
//	"Whenever you draw a card, this creature gets +2/+2 until end of turn."
//
// WheneverYouDraw, once per card drawn, with a layer 7c boost on the
// source that ends at cleanup (thisCreatureUntilEOT).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95a08f25-8d44-4704-a4d1-a098bf010932",
		Name:         "Burlfist Oak",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Burlfist Oak — this creature gets +2/+2 until end of turn",
				effects.ThisCreatureUntilEOT("Burlfist Oak — +2/+2 until end of turn", 2, 2)),
		},
	})
}
