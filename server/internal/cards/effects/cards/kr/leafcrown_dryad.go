package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "01914751-af8e-4aea-a254-6fa4f9fd3457",
		Name:            "Leafcrown Dryad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{3}{G}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("reach"),
		},
	})
}
