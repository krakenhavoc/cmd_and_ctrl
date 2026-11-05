package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tenacious Tosk — Creature — Tosk Rogue {3}{G}, 4/4:
//
//	"This creature can't be blocked by more than one creature.
//	 Whenever this creature deals combat damage to a player, put a
//	 +1/+1 counter on it."
//
// The block limit is Professional Wrestler's CR 509.1b MaxBlockers;
// the counter is put on the Tosk only if it is still the same object
// (CounterOnThis, CR 400.7).
//
// No simplification. Star Trek Commander (#4974).
func init() {
	Register(Spec{
		OracleID:     "48c2dc82-7bed-4730-8b8b-1970b743414f",
		Name:         "Tenacious Tosk",
		Completeness: CompletenessFull,
		BlockRules:   []game.BlockRule{MaxBlockers(OnSelf(), 1)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Tenacious Tosk — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
