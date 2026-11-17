package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcbound Slith — Artifact Creature — Slith {2}, 0/0 (Darksteel, #5605):
//
//	"Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it.
//	 Modular 1 (This creature enters with a +1/+1 counter on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)"
//
// The counter goes on only if the Slith is still on the battlefield as
// the trigger resolves. Modular is the engine's keyword (#2012,
// CR 702.43a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fa113da-44aa-4c84-8f63-b1cad63da926",
		Name:            "Arcbound Slith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"modular 1"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Arcbound Slith — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
