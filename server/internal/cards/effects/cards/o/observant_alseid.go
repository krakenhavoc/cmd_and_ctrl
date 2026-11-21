package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Observant Alseid — Enchantment Creature — Nymph {2}{W}, 2/2 (Theros, #5561):
//
//	"Bestow {4}{W} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	 Vigilance
//	 Enchanted creature gets +2/+2 and has vigilance."
//
// Bestow (ADR 0141, #2862): the creature, or the same keyword and bonus
// on a creature that already has a body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0eb0c9d2-bc5a-4ba3-bd83-57af1da51c3b",
		Name:            "Observant Alseid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{4}{W}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("vigilance"),
		},
	})
}
