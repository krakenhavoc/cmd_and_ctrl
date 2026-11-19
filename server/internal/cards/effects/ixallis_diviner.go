package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ixalli's Diviner — Creature — Human Druid {1}{G}, 0/3 (Foundations
// Jumpstart, BNS batch 32):
//
//	"When this creature enters, it explores."
//
// Cenote Scout's body: the explorer is the object that entered
// (thatCreatureExplores), so a Diviner that has left by resolution still
// reveals the card but gets no counter (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56c930c4-77e8-4547-8d55-663dc68ccf13",
		Name:         "Ixalli's Diviner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ixalli's Diviner — it explores", thatCreatureExplores),
		},
	})
}
