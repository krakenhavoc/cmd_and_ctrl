package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyship Plunderer — Creature — Human Pirate {1}{U}, 2/1:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, for each kind
//	 of counter on target permanent or player, give that permanent or
//	 player another counter of that kind."
//
// Powerful Broker's body (mshA6AnotherOfEachKind) behind a combat-damage
// trigger: not proliferate, so the target is the trigger's and not a
// choice at resolution. The target is chosen as the trigger goes on the
// stack, and the kinds are read as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c5e47644-c72c-40f9-b65f-57fcea075466",
		Name:            "Skyship Plunderer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisDealsCombatDamageToAPlayer("Skyship Plunderer — for each kind of counter on target permanent or player, give it another counter of that kind",
					mshA6AnotherOfEachKindOnTarget),
				ardennTargetPermanentOrPlayer()),
		},
	})
}
