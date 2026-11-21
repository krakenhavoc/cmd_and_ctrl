package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meriadoc Brandybuck — Legendary Creature — Halfling Citizen {1}{G}, 2/2
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever one or more Halflings you control attack a player, create a
//	 Food token."
//
// Attuma's "attack a player" guard (CR 603.2c): one declaration collapses
// to a Food per player attacked, however many Halflings go at that player,
// and a second player attacked is a second Food. Meriadoc is a Halfling
// himself. "Halfling" is read through the source's text, so a
// text-changing effect retypes the trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fec6a4a7-4ded-4062-b99b-c1826dea7113",
		Name:         "Meriadoc Brandybuck",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, gapB125TribeAttackAPlayer("Halfling"),
				"Meriadoc Brandybuck — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1}))),
		},
	})
}
