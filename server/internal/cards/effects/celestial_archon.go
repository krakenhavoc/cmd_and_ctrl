package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celestial Archon — Enchantment Creature — Archon {3}{W}{W}, 4/4
// (EDHREC rank 13032):
//
//	"Bestow {5}{W}{W} (If you cast this card for its bestow cost, it's
//	 an Aura spell with enchant creature. It becomes a creature again
//	 if it's not attached.)
//	 Flying, first strike
//	 Enchanted creature gets +4/+4 and has flying and first strike."
//
// A 4/4 flier with first strike, or the same again stapled onto a
// creature that already has a body, and a 4/4 flier left behind when
// that creature is answered (ADR 0141, #2862).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d2b29bb0-9fb5-46cd-a546-fa6c4ff6113c",
		Name:            "Celestial Archon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{5}{W}{W}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(4, 4),
			GrantToAttached("flying", "first strike"),
		},
	})
}
