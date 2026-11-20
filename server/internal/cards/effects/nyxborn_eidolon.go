package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyxborn Eidolon — Enchantment Creature — Spirit {1}{B}, 2/1 (Born of the Gods,
// #5559):
//
//	"Bestow {4}{B} (If you cast this card for its bestow cost, it's an
//	 Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Enchanted creature gets +2/+1."
//
// Nyxborn Rollicker's shape (ADR 0141, #2862): the body, or the same
// bonus stapled onto a creature, and the body left behind when that
// creature is answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd7fd82f-154a-4001-bb8f-7fb50ac77b6b",
		Name:         "Nyxborn Eidolon",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{4}{B}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
		},
	})
}
