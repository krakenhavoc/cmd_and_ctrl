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
// Same declared auto-tap simplification as Wild Growth: the planner
// does not count the extra mana, so it may tap one land too many and
// the surplus floats (ADR 0074 §7).
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
