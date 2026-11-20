package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyxborn Wolf — Enchantment Creature — Wolf {2}{G}, 3/1 (Born of the Gods,
// #5559):
//
//	"Bestow {4}{G} (If you cast this card for its bestow cost, it's an
//	 Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Enchanted creature gets +3/+1."
//
// Nyxborn Rollicker's shape (ADR 0141, #2862): the body, or the same
// bonus stapled onto a creature, and the body left behind when that
// creature is answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fa11a6a-a760-45db-8341-f83a0f69c17f",
		Name:         "Nyxborn Wolf",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{4}{G}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(3, 1),
		},
	})
}
