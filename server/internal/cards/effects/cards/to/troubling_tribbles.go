package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Troubling Tribbles — Creature — Tribble {2}{W}, 0/1 (Star Trek):
//
//	This creature can't block.
//	At the beginning of your upkeep, create a token that's a copy of
//	this creature.
//
// The copy is CreateTokenCopy over the source, which reads its copiable
// values off the permanent, or its last-known ones once it has left
// (CR 608.2h), so a Tribble that dies with the trigger on the stack still
// breeds. Every copy carries both abilities (CR 707.2), so the herd
// doubles each upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc76eddd-8f2a-45e3-95d2-3074d971effb",
		Name:         "Troubling Tribbles",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Troubling Tribbles — create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
