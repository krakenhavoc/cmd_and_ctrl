package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wingsteed Rider — Creature — Human Knight {1}{W}{W}, 2/2:
//
//	"Flying
//	 Heroic — Whenever you cast a spell that targets this creature, put a
//	 +1/+1 counter on this creature."
//
// Heroic (CR 207.2c) is the shared `heroic` cast trigger; the counter
// goes only on the same object (CR 400.7), as for Favored Hoplite.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "9c46adaa-ab39-4c0a-99bd-935b6c4f732f",
		Name:            "Wingsteed Rider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			heroic("Wingsteed Rider — a +1/+1 counter on it (heroic)", thisGetsAPlusOneCounter),
		},
	})
}
