package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyxborn Triton — Enchantment Creature — Merfolk {2}{U}, 2/3 (Born of the Gods,
// #5559):
//
//	"Bestow {4}{U} (If you cast this card for its bestow cost, it's an
//	 Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Enchanted creature gets +2/+3."
//
// Nyxborn Rollicker's shape (ADR 0141, #2862): the body, or the same
// bonus stapled onto a creature, and the body left behind when that
// creature is answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdd37bbf-09db-4c1f-9947-312de48a0f4c",
		Name:         "Nyxborn Triton",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{4}{U}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(2, 3),
		},
	})
}
