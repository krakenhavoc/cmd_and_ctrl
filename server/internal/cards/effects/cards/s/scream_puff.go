package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scream Puff — Creature — Horror {4}{B}, 4/5 (Wilds of Eldraine, #4471):
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, create a Food
//	 token."
//
// Deathtouch is the engine's keyword (CR 702.2). The trigger is the shared
// "deals combat damage to a player" shape, and the Food is the token
// table's, created under the trigger's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b46ddf4b-7ed8-49c4-8828-f0fbbc023c1f",
		Name:            "Scream Puff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Scream Puff — create a Food token",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
