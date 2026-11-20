package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War-Wing Siren — Creature — Siren Soldier {2}{U}, 1/3 (Journey into Nyx, #5558):
//
//	"Flying
//	 Heroic — Whenever you cast a spell that targets this creature, put
//	 a +1/+1 counter on this creature."
//
// Wingsteed Rider's heroic counter on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0e49ec18-f979-4168-b7df-7728f3fe442e",
		Name:            "War-Wing Siren",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			heroic("War-Wing Siren — a +1/+1 counter on it (heroic)", thisGetsAPlusOneCounter),
		},
	})
}
