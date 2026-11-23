package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "e91053ae-682a-4bbe-8394-fd65e626bacc",
		Name:            "Stronghold Rats",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Stronghold Rats — each player discards a card",
				effects.TspB1EachPlayerDiscardsACard("Stronghold Rats — discard a card")),
		},
	})
}
