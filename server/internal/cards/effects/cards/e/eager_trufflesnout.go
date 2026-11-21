package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eager Trufflesnout — Creature — Boar {2}{G}, 4/2 (Foundations, #4465):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, create a
//	 Food token."
//
// Scream Puff's trigger on a trampler: damage trampled over to a player
// counts, and the trigger fires once per combat damage step in which the
// Boar deals combat damage to a player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7f2ba562-8d5e-4915-967e-54916c3c5876",
		Name:            "Eager Trufflesnout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Eager Trufflesnout — create a Food token",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
