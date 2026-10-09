package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyxborn Rollicker — Enchantment Creature — Satyr {R}, 1/1 (EDHREC
// rank 20800):
//
//	"Bestow {1}{R} (If you cast this card for its bestow cost, it's an
//	 Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Enchanted creature gets +1/+1."
//
// The plainest bestow card there is (ADR 0141, #2862): a 1/1 for one,
// or +1/+1 for two that leaves the 1/1 behind.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebf2974a-963e-425a-8f8c-55f0a36984c3",
		Name:         "Nyxborn Rollicker",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{1}{R}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
		},
	})
}
