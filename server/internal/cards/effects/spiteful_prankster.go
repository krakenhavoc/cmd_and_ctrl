package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiteful Prankster — Creature — Devil {2}{R}, 3/2 (Jumpstart 2022, #5948):
//
//	"During your turn, this creature has first strike.
//	 Whenever another creature dies, this creature deals 1 damage to target
//	 player or planeswalker."
//
// Feisty Spikeling's turn-gated first strike (a layer-6 grant while its
// controller is the active player) and Deathgreeter's any-creature dies
// watcher; the Prankster deals the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58cbac5c-d58e-4794-87f3-9563fce1b53f",
		Name:         "Spiteful Prankster",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{KeywordGrant(gapB158ThisOnYourTurn, "first strike")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventLTB, AnotherCreatureDied, "Spiteful Prankster — 1 damage to target player or planeswalker",
					sourceDealsDamageToEachLegalTarget(1)),
				targetPlayerOrPlaneswalker()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
