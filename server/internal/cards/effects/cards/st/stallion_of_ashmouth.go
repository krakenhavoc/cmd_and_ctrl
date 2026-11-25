package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stallion of Ashmouth — Creature — Nightmare Horse {3}{B}, 3/3
// (Shadows over Innistrad, #5543):
//
//	"Delirium — {1}{B}: This creature gets +1/+1 until end of turn.
//	 Activate only if there are four or more card types among cards in
//	 your graveyard."
//
// Delirium is an ability word (CR 207.2c), here an activation condition
// (CR 602.5b): the engine reads the activator's own graveyard each time the
// ability is offered and again as it is activated, and the pump is
// Shivan Dragon's shared self-pump row.
//
// No simplification.
func init() {
	pump := effects.SelfPumpRow("Stallion of Ashmouth", "{1}{B}", 1, 1)
	pump.Label = "Delirium — {1}{B}: This creature gets +1/+1 until end of turn. Activate only if there are four or more card types among cards in your graveyard."
	pump.Condition = func(g *game.Game, controller, _ uuid.UUID) bool {
		return effects.DrcDeliriumMet(g, controller)
	}
	effects.Register(effects.Spec{
		OracleID:     "0dc6a90e-c517-445e-9a93-29b3119fb4dc",
		Name:         "Stallion of Ashmouth",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{pump},
	})
}
