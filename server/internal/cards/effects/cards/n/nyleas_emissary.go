package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nylea's Emissary — Enchantment Creature — Cat {3}{G}, 3/3 (Theros, #5561):
//
//	"Bestow {5}{G} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	 Trample
//	 Enchanted creature gets +3/+3 and has trample."
//
// Bestow (ADR 0141, #2862): the creature, or the same keyword and bonus
// on a creature that already has a body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec30e221-ffc8-4867-ba9c-f39712b3aa93",
		Name:            "Nylea's Emissary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{5}{G}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("trample"),
		},
	})
}
