package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reaper of Flight Moonsilver — Creature — Angel {3}{W}{W}, 3/3 (Shadows
// over Innistrad, #5543):
//
//	"Flying
//	 Delirium — Sacrifice another creature: This creature gets +2/+1 until
//	 end of turn. Activate only if there are four or more card types among
//	 cards in your graveyard."
//
// Delirium is an activation condition (CR 602.5b) on the activator's own
// graveyard, read each time the ability is offered and as it is activated.
// The sacrifice is the whole cost, paid as the ability is announced
// (CR 602.2c); "another" is object identity, so the Angel can't pay with
// itself. The boost lands on the Angel only if it is still the same object
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94cce37a-a752-4cde-a5b5-e8a15f8af52c",
		Name:            "Reaper of Flight Moonsilver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Delirium — Sacrifice another creature: This creature gets +2/+1 until end of turn. Activate only if there are four or more card types among cards in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet, Pump: &game.Pump{Power: 2, Toughness: 1}},
			Cost: game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature", effects.Creature())),
			},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.DrcDeliriumMet(g, controller)
			},
			Effect: effects.ThisGetsUntilEndOfTurn(2, 1, "Reaper of Flight Moonsilver — +2/+1"),
		}},
	})
}
