package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morinfen — Legendary Creature — Phyrexian Horror {3}{B}{B}, 5/4 (Weatherlight, #5629):
//
//	"Flying
//	 Cumulative upkeep—Pay 1 life. (At the beginning of your upkeep, put an
//	 age counter on this permanent, then sacrifice it unless you pay its
//	 upkeep cost for each age counter on it.)"
//
// Gallowbraid's cumulative upkeep with flying in place of trample: at N
// age counters the controller pays N life or sacrifices it (CR 702.24a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d69f3aaf-8462-432d-afaf-4176ac04962b",
		Name:            "Morinfen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeepPaying("Morinfen — cumulative upkeep: pay 1 life", LifePayment(1), "", ""),
		},
	})
}
