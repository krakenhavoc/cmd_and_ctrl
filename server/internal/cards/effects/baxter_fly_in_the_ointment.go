package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baxter, Fly in the Ointment — {3}{U} Legendary Creature — Insect
// Mutant Scientist 2/2 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Whenever Baxter enters or attacks, each creature you control with a
//	 counter on it gains flying until end of turn.
//	 Whenever you draw a card, put a +1/+1 counter on Baxter."
//
// Two triggers. The first is one ability with two conditions (Sun
// Titan's WhenThisEntersOrAttacks). The creatures that gain flying are
// the ones its controller controls with a counter of any kind on them
// as it resolves, and that set is locked then (CR 611.2c): a creature
// that gets its first counter later that turn does not fly, and one
// that loses its counters keeps flying. The second is one trigger per
// card drawn, the Oneirophage body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94551fbd-a72c-4a1f-be01-7ba1c7c03eed",
		Name:         "Baxter, Fly in the Ointment",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Baxter, Fly in the Ointment — each creature you control with a counter on it gains flying until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl(), b41HasACounter()),
					Keywords: []string{"flying"},
					Label:    "Baxter, Fly in the Ointment — flying",
				})),
			WheneverYouDraw("Baxter, Fly in the Ointment — put a +1/+1 counter on Baxter",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
