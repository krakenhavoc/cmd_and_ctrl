package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pathfinding Axejaw — Creature — Dinosaur {3}{G}, 4/3 (The Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, it explores."
//
// Cenote Scout's trigger: the explorer is the object that entered, read
// off the trigger (thatCreatureExplores). A land on top goes to the hand,
// anything else gives the Axejaw a +1/+1 counter and asks whether to put
// the card into the graveyard (CR 701.44a). An Axejaw that has left by
// the time the trigger resolves still explores by its last known
// information: the card is revealed, and no counter is put on anything
// (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54452df0-560c-4561-9dd2-3903ae49c91b",
		Name:         "Pathfinding Axejaw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pathfinding Axejaw — it explores", thatCreatureExplores),
		},
	})
}
