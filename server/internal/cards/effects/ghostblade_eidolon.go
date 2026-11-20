package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghostblade Eidolon — Enchantment Creature — Spirit {2}{W}, 1/1 (Born of the Gods,
// #5559):
//
//	"Bestow {5}{W} (If you cast this card for its bestow cost, it's an
//	 Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Double strike (This creature deals both first-strike and regular
//	 combat damage.)
//	 Enchanted creature gets +1/+1 and has double strike."
//
// Celestial Archon's shape (ADR 0141): the keyword printed on the body
// and granted to the host while bestowed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "78a1f291-22ed-41f1-87fd-0bdc97f9c065",
		Name:            "Ghostblade Eidolon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{5}{W}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("double strike"),
		},
	})
}
