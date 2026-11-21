package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nyxborn Shieldmate — Enchantment Creature — Human Soldier {W}, 1/2 (Born of the Gods,
// #5559):
//
//	"Bestow {2}{W} (If you cast this card for its bestow cost, it's an
//	 Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Enchanted creature gets +1/+2."
//
// Nyxborn Rollicker's shape (ADR 0141, #2862): the body, or the same
// bonus stapled onto a creature, and the body left behind when that
// creature is answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1012f33-7ebd-47c6-b670-1b0dc046fbce",
		Name:         "Nyxborn Shieldmate",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{2}{W}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
		},
	})
}
