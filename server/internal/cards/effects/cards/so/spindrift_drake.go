package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spindrift Drake — Creature — Drake {U}, 2/1 (Stronghold, S82 OLD b749, #5627):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice this creature unless you pay
//	 {U}."
//
// Darba's upkeep pay-or-sacrifice with a fixed {U}: the controller is asked as
// the trigger resolves, and the table waits in the upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53599f89-1598-4d3d-8aa7-75b397b258eb",
		Name:            "Spindrift Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Spindrift Drake — sacrifice this creature unless you pay {U}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{U}",
						Question: "Spindrift Drake — pay {U} or sacrifice Spindrift Drake?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
