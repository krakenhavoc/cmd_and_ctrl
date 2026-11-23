package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Purphoros's Emissary — Enchantment Creature — Ox {3}{R}, 3/3 (Born of
// the Gods):
//
//	"Bestow {6}{R} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Menace
//	 Enchanted creature gets +3/+3 and has menace."
//
// Spearpoint Oread's shape (ADR 0141, #2862): the creature, or the same
// bonus on a creature that already has a body, and menace on both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d78d7fd6-778b-4030-a776-29ddce83e33f",
		Name:            "Purphoros's Emissary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{6}{R}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("menace"),
		},
	})
}
