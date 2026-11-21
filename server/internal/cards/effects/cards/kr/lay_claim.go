package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lay Claim — Enchantment — Aura {5}{U}{U} (Jumpstart 2022, #5948):
//
//	"Enchant permanent
//	 You control enchanted permanent.
//	 Cycling {2}"
//
// Control Magic's layer-2 static on the widest enchant clause (a creature, a
// land, anything), plus the cycling keyword's own hand ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7111365-b96b-4d6a-8c3c-73a0d8baf641",
		Name:         "Lay Claim",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
