package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monk of the Open Hand — Creature — Elf Monk {W}, 1/1 (Adventures in the
// Forgotten Realms):
//
//	"Flurry of Blows — Whenever you cast your second spell each turn, put a
//	 +1/+1 counter on this creature."
//
// Flurry of Blows is a flavor word (no rules meaning). The trigger is
// YouCastYourSecondSpellEachTurn, as Devoted Duelist reads it: the cast path
// bumps the per-turn tally before it emits EventCast, so a total of two is
// the second spell, and a spell cast before the Monk entered still counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d3998fc-a68a-4923-a456-b3e43ab6decb",
		Name:         "Monk of the Open Hand",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Monk of the Open Hand — flurry of blows: put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
