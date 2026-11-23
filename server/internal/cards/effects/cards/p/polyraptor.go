package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Polyraptor — Creature — Dinosaur {6}{G}{G}, 5/5 (Special Guests):
//
//	"Enrage — Whenever this creature is dealt damage, create a token that's
//	 a copy of this creature."
//
// One trigger per batch of damage (CR 120.3). The token is a copy of the
// Polyraptor's copiable values as it last existed (CR 707.2, CR 608.2h), so
// it carries the same trigger and copies again when it is damaged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e47e3a40-f51f-41ed-8e7c-06200a2abc22",
		Name:         "Polyraptor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Polyraptor — Enrage: create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
