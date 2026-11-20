package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leafcrown Dryad — Enchantment Creature — Nymph Dryad {1}{G}, 2/2 (Theros, #5561):
//
//	"Bestow {3}{G} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	 Reach
//	 Enchanted creature gets +2/+2 and has reach."
//
// Bestow (ADR 0141, #2862): the creature, or the same keyword and bonus
// on a creature that already has a body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01914751-af8e-4aea-a254-6fa4f9fd3457",
		Name:            "Leafcrown Dryad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{3}{G}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("reach"),
		},
	})
}
