package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spellweaver Eternal — Creature — Zombie Snake Wizard {1}{U}, 2/1
// (Hour of Devastation, #5534):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 Afflict 2 (Whenever this creature becomes blocked, defending player
//	 loses 2 life.)"
//
// Prowess is a printed keyword (CR 702.108a).
// Afflict is the shared Afflict row (CR 702.130a): when this creature
// becomes blocked, the defending player loses the life, which is life
// loss and not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "71f3d3d3-3f82-43f6-b2a9-d21afbdd7739",
		Name:            "Spellweaver Eternal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered:       []game.TriggeredAbility{Afflict(2, "Spellweaver Eternal")},
	})
}
