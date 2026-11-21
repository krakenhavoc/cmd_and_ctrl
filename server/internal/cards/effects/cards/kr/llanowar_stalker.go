package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Stalker — Creature — Elf Warrior {G}, 1/1 (Dominaria United):
//
//	"Whenever another creature you control enters, this creature gets +1/+0
//	 until end of turn."
//
// Champion of Lambholt's "another creature you control enters" trigger with
// a self pump until end of turn (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90d375fa-0487-4990-9a98-260f9ee28d2a",
		Name:         "Llanowar Stalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl("Llanowar Stalker — +1/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 0, "Llanowar Stalker — +1/+0 until end of turn")),
		},
	})
}
