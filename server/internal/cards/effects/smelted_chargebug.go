package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smelted Chargebug — Artifact Creature — Insect {1}{R}, 1/3:
//
//	"Menace
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}. If you do, another
//	 target attacking creature gets +1/+0 and gains menace until end of
//	 turn."
//
// ADR 0129 §3 (#1995): Consul's Shieldguard's shape with +1/+0 and
// menace.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8bc30bd5-554e-4fcf-9dec-c7634921858a",
		Name:            "Smelted Chargebug",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Smelted Chargebug", 2),
			anotherAttackerMayPayEnergy("Smelted Chargebug", "give it +1/+0 and menace",
				firstTargetGainsUntilEndOfTurn(1, "Smelted Chargebug — +1/+0 and menace until end of turn", "menace")),
		},
	})
}
