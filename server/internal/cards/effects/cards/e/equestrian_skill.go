package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Equestrian Skill — Enchantment — Aura {3}{G} (Shadows over Innistrad, S82
// OLD):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 As long as enchanted creature is a Human, it has trample."
//
// A flat pump (a layer 7c modify) and Champion's Helm's conditional keyword
// read on every recompute.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c8c12ff-24ff-4279-9bde-da8899d2c2d5",
		Name:         "Equestrian Skill",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttachedWhile(effects.OldB289HostIsHuman, "trample"),
		},
	})
}
