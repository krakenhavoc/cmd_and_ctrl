package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "7f2ba562-8d5e-4915-967e-54916c3c5876",
		Name:            "Eager Trufflesnout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Eager Trufflesnout — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
