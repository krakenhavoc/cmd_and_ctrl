package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spearpoint Oread — Enchantment Creature — Nymph {2}{R}, 2/2 (Theros, #5561):
//
//	"Bestow {5}{R} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	 First strike
//	 Enchanted creature gets +2/+2 and has first strike."
//
// Bestow (ADR 0141, #2862): the creature, or the same keyword and bonus
// on a creature that already has a body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2fa7bb26-18eb-4ff0-8558-f6f3981a84ce",
		Name:            "Spearpoint Oread",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{5}{R}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("first strike"),
		},
	})
}
