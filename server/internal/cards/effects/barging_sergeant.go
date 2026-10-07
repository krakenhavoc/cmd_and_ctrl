package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barging Sergeant — Creature — Minotaur Soldier {4}{R}, 4/2:
//
//	"Haste
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)"
//
// Mentor (CR 702.136) is the shared Mentor() trigger; see
// hammer_dropper.go for what its source-relative clause guarantees.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "af422a4b-087c-4753-9288-c31fb3457740",
		Name:            "Barging Sergeant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{Mentor("Barging Sergeant — mentor")},
	})
}
