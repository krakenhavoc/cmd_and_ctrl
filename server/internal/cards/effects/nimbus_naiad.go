package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nimbus Naiad — Enchantment Creature — Nymph {2}{U}, 2/2 (Theros, #5561):
//
//	"Bestow {4}{U} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	 Flying
//	 Enchanted creature gets +2/+2 and has flying."
//
// Bestow (ADR 0141, #2862): the creature, or the same keyword and bonus
// on a creature that already has a body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1b3b14b9-088d-48df-8d21-54ee68989ac0",
		Name:            "Nimbus Naiad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{4}{U}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying"),
		},
	})
}
