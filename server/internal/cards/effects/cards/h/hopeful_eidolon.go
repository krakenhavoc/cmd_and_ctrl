package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hopeful Eidolon — Enchantment Creature — Spirit {W}, 1/1 (EDHREC
// rank 15226):
//
//	"Bestow {3}{W} (If you cast this card for its bestow cost, it's an
//	 Aura spell with enchant creature. It becomes a creature again if
//	 it's not attached.)
//	 Lifelink
//	 Enchanted creature gets +1/+1 and has lifelink."
//
// A one-drop lifelinker, or lifelink on the creature that hits hardest
// (ADR 0141, #2862).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13c1ca1e-cd17-49ef-bc31-97bff99ec1e3",
		Name:            "Hopeful Eidolon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{3}{W}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("lifelink"),
		},
	})
}
