package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protective Bubble — Enchantment — Aura {3}{U} (Lorwyn, S82 OLD, #5593):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked and has shroud."
//
// Cloak of Mists' restriction and a shroud grant, both following the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1fba543c-bd22-4f7b-9ff4-0631af8ee814",
		Name:         "Protective Bubble",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantBeBlocked),
			effects.GrantToAttached("shroud"),
		},
	})
}
