package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandskitter Outrider — Creature — Goblin Soldier {3}{B}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 When this creature enters, it endures 2. (Put two +1/+1 counters on
//	 it or create a 2/2 white Spirit creature token.)"
//
// Endure is endure_n.go (CR 701.63a): its controller chooses the counters
// or the Spirit as the trigger resolves. An Outrider that has left the
// battlefield by then can only make the Spirit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eae480fe-d9fe-4dbe-8b78-b65f398b0af3",
		Name:            "Sandskitter Outrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sandskitter Outrider — it endures 2", ThisEndures("Sandskitter Outrider", 2)),
		},
	})
}
