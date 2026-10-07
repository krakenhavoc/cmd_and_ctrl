package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eddytrail Hawk — Creature — Bird {1}{W}, 1/2:
//
//	"Flying
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}. If you do, another
//	 target attacking creature gains flying until end of turn."
//
// ADR 0129 §3 (#1995): Consul's Shieldguard's shape with flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "748729bf-bfa0-4cbf-b643-ed5d7d007081",
		Name:            "Eddytrail Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Eddytrail Hawk", 2),
			anotherAttackerMayPayEnergy("Eddytrail Hawk", "give it flying",
				firstTargetGainsUntilEndOfTurn(0, "Eddytrail Hawk — flying until end of turn", "flying")),
		},
	})
}
