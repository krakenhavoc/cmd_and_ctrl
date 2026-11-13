package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallowbraid — Legendary Creature — Phyrexian Horror {3}{B}{B}, 5/5 (Weatherlight, #5629):
//
//	"Trample
//	 Cumulative upkeep—Pay 1 life. (At the beginning of your upkeep, put an
//	 age counter on this permanent, then sacrifice it unless you pay its
//	 upkeep cost for each age counter on it.)"
//
// Trample is the printed keyword. The cumulative upkeep is
// CumulativeUpkeepPaying over LifePayment (CR 702.24a): at N age counters
// the controller pays N life or sacrifices it, and can pay only with at
// least that much life (CR 119.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4a4418fa-7574-4027-aa09-4ca789be15ba",
		Name:            "Gallowbraid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeepPaying("Gallowbraid — cumulative upkeep: pay 1 life", LifePayment(1), "", ""),
		},
	})
}
