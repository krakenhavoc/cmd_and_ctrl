package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ophidian — Creature — Snake {2}{U}, 1/3:
//
//	"Whenever this creature attacks and isn't blocked, you may draw a card. If you do, this
//	 creature assigns no combat damage this turn."
//
// #3531: the draw is the "if you do".
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "953bdefa-4639-4524-af5c-5055d1289f40",
		Name:         "Ophidian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			unblockedAttackerMayDraw("Ophidian", "draw a card", 1, func(ctx *Context, _ uuid.UUID) error {
				return DrawCards{N: 1}.Apply(ctx)
			}),
		},
	})
}
