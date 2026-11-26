package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit Away — {5}{U}{U} Aura: "Enchant creature. You control enchanted
// creature. Enchanted creature gets +2/+2 and has flying."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15b618f0-281f-4745-8504-bd70e7b4a267",
		Name:         "Spirit Away",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
		},
	})
}
