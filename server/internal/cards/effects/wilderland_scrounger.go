package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wilderland Scrounger — Creature — Wolf {4}{G}, 3/6 (The Hobbit):
//
//	"Ferocious — Whenever this creature attacks while you control a
//	 creature with power 4 or greater, put a +1/+1 counter on each
//	 creature you control."
//
// See hobbitFerociousAttack for the "while" window. The counters use the
// shared "each creature you control" body (Archangel of Thune).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a2946ad-7a7b-4cda-8591-da8c7da5e072",
		Name:         "Wilderland Scrounger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, hobbitFerociousAttack, "Wilderland Scrounger — ferocious: a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
	})
}
