package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boar-q-pine — Creature — Boar Porcupine {2}{R}, 2/2:
//
//	"Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	 creature."
//
// Chandra's Emberling's trigger without the haste: it fires on the cast
// (CR 601.2i), so a countered spell still grows it, and a Boar-q-pine
// that has left the battlefield gets nothing.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "551e4356-5b1f-475c-a64b-9c50ea9b0368",
		Name:         "Boar-q-pine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Boar-q-pine — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
