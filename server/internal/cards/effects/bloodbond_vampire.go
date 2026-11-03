package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodbond Vampire — Creature — Vampire Shaman Ally {2}{B}{B}, 3/3:
//
//	"Whenever you gain life, put a +1/+1 counter on this creature."
//
// Ajani's Pridemate's trigger on a four-mana body: once per life-gain
// event, not once per point, and nothing for a Vampire that has left the
// battlefield before it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2057118-9fd1-431d-9441-fef198d67ff2",
		Name:         "Bloodbond Vampire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Bloodbond Vampire — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
