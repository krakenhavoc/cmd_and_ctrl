package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akoum Firebird — Creature — Phoenix {2}{R}{R}, 3/3 (Battle for Zendikar):
//
//	"Flying, haste
//	 This creature attacks each combat if able.
//	 Landfall — Whenever a land you control enters, you may pay {4}{R}{R}. If
//	 you do, return this card from your graveyard to the battlefield."
//
// The landfall trigger watches from the graveyard, the only zone it can fire
// from; the payment is offered at resolution and only for the same card
// still in the graveyard (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10ab3102-43a3-4c60-8641-2c5a7c1e29c2",
		Name:            "Akoum Firebird",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.Landfall("Akoum Firebird — you may pay {4}{R}{R} to return it to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayForGraveyardReturn(g, item, "{4}{R}{R}", "Akoum Firebird — pay {4}{R}{R} to return it to the battlefield?")
				})),
		},
	})
}
