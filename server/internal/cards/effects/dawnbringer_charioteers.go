package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnbringer Charioteers — Creature — Human Soldier {2}{W}{W}, 2/4 (Journey into Nyx, #5558):
//
//	"Flying, lifelink
//	 Heroic — Whenever you cast a spell that targets this creature, put
//	 a +1/+1 counter on this creature."
//
// Wingsteed Rider's heroic counter, with lifelink beside the flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "29263d48-cb28-478b-a99e-5a18dbc924f0",
		Name:            "Dawnbringer Charioteers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			heroic("Dawnbringer Charioteers — a +1/+1 counter on it (heroic)", thisGetsAPlusOneCounter),
		},
	})
}
