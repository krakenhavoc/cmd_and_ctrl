package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodmad Vampire — Creature — Vampire Berserker {2}{R}, 4/1 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it.
//	 Madness {1}{R}"
//
// A plain combat-damage trigger (the counter goes on only while it is
// still on the battlefield) and madness as the shared keyword
// (Spec.Madness).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8a20fc7-025b-403a-b893-ff5efcbdc9d8",
		Name:         "Bloodmad Vampire",
		Completeness: CompletenessFull,
		Madness:      "{1}{R}",
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Bloodmad Vampire — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
