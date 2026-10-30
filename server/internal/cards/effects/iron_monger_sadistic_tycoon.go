package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Monger, Sadistic Tycoon — Legendary Artifact Creature — Human
// Villain {2}{B}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever a creature you control connives, put a +1/+1 counter on
//	 each Villain you control."
//
// The trigger watches game.EventConnived (#4641, CR 701.50f), emitted once
// per creature as it finishes conniving. Two creatures conniving together
// connive one after another (CR 701.50c), so it triggers twice. It counts
// a creature that left the battlefield before its turn to connive
// (CR 701.50b, last controller). The counters are Citizen V's body: every
// Villain the controller controls when the ability resolves, Iron Monger
// himself among them, each its own placement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7324d8c0-3a30-452b-b05d-ec7b6a98666c",
		Name:            "Iron Monger, Sadistic Tycoon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlConnives("Iron Monger, Sadistic Tycoon — put a +1/+1 counter on each Villain you control",
				Do(Instruction(msc5CounterOnEachVillain))),
		},
	})
}
