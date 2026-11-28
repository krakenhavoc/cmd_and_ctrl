package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aluren — Enchantment {2}{G}{G} (Tempest, S82 OLD batch 765, #5628):
//
//	"Any player may cast creature spells with mana value 3 or less without
//	 paying their mana costs and as though they had flash."
//
// A granted alternative cost (CR 118.9) made to EVERY player's casts, not only
// its controller's (Weftwalking's EveryPlayer), over creature spells with mana
// value 3 or less, from any zone the printed mana cost could be paid. The free
// cast is the offer's: a creature cast this way may be cast as though it had
// flash, and the same creature cast for its printed cost keeps its own timing
// (CR 601.3c). X is 0 on a free cast (CR 107.3b), and additional costs are
// still paid (CR 118.9d). The offer is derived from the battlefield on every
// cast query, so it ends when Aluren does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                "44e02e70-07ab-4795-8037-55185bd11d54",
		Name:                    "Aluren",
		Completeness:            effects.CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.AnyPlayerMayCastSmallCreaturesFreeWithFlash()},
	})
}
