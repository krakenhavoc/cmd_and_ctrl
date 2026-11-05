package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phlox, Unorthodox Doctor — Legendary Creature — Denobulan Doctor
// {2}{G}{W}, 3/4 (Star Trek, #4973):
//
//	"Trample
//	 Whenever a creature you control with a counter on it deals combat
//	 damage to a player, draw a card."
//
// Ray Fillet, Wave Warrior's trigger: one trigger per creature that
// connects, the dealer's counters read as the damage is dealt
// (yourCreatureWithACounterDealtCombatDamageToAPlayer). "A counter" is any
// kind (CR 122.1). Trample is the printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b04cb333-ca7a-464f-a084-03cf8f468baf",
		Name:            "Phlox, Unorthodox Doctor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, yourCreatureWithACounterDealtCombatDamageToAPlayer,
				"Phlox, Unorthodox Doctor — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
