package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stronghold Rats — Creature — Rat {2}{B}, 2/1 (Future Sight, #5594):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 Whenever this creature deals combat damage to a player, each player
//	 discards a card."
//
// Shadow is the engine's; the trigger is the shared combat-damage
// constructor with Mindlash Sliver's "each player discards a card" body:
// every player still in the game, the Rats' controller included, picks a
// card of their own, in APNAP order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e91053ae-682a-4bbe-8394-fd65e626bacc",
		Name:            "Stronghold Rats",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Stronghold Rats — each player discards a card",
				tspB1EachPlayerDiscardsACard("Stronghold Rats — discard a card")),
		},
	})
}
