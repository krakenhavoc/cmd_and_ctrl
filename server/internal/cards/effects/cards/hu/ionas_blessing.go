package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iona's Blessing — Enchantment — Aura {3}{W}:
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has vigilance, and can block an
//	 additional creature each combat."
//
// Three ordinary Aura statics over AttachedToSource: PumpAttached for
// the +2/+2 (layer 7c), GrantToAttached for vigilance (layer 6), and
// CanBlockAdditional (#1706, also layer 6) for the block line. The
// Aura attaches at resolution the way every "Enchant creature" clause
// does (ADR 0036 decision 5) — no OnResolve needed, per Angelic Gift.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc5bde87-231e-424a-a8fa-cd9910a09d43",
		Name:         "Iona's Blessing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("vigilance"),
			effects.CanBlockAdditional(effects.AttachedToSource, 1),
		},
	})
}
