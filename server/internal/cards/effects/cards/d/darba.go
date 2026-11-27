package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darba — Creature — Bird Beast {3}{G}, 5/4 (Prophecy, S82 OLD b671, #5617):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you pay
//	 {G}{G}."
//
// Fettergeist's upkeep pay-or-sacrifice with a fixed {G}{G}: the controller is
// asked as the trigger resolves, and the table waits in the upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c64c0cb5-5a8c-4a2f-a0fc-8b080aaf90bb",
		Name:         "Darba",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Darba — sacrifice this creature unless you pay {G}{G}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{G}{G}",
						Question: "Darba — pay {G}{G} or sacrifice Darba?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
