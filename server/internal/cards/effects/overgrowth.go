package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overgrowth — Enchantment — Aura {2}{G} (#391):
//
//	"Enchant land
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional {G}{G}."
//
// Wild Growth for two more mana and one more {G}. It is the same
// CR 605.1b triggered mana ability and the same constructor; what it
// proves is that the produced string is the ordinary
// ParseProducedMana grammar, so "more than one token from one trigger"
// needed nothing new.
//
// The auto-tap planner counts both extra {G} (ADR 0074, amendment of
// 2026-10-30): a Forest under Overgrowth plans as {G}{G}{G}.
//
// No simplification of the card itself.
func init() {
	Register(Spec{
		OracleID:     "e6ccebf4-f4f0-404f-a634-4d751ef9c8aa",
		Name:         "Overgrowth",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForMana("Overgrowth — add an additional {G}{G}", "{G}{G}"),
		},
	})
}
