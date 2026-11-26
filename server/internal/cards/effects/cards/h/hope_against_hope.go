package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hope Against Hope — Enchantment — Aura {2}{W} (Shadows over Innistrad, S82
// OLD):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each creature you control.
//	 As long as enchanted creature is a Human, it has first strike."
//
// Adaptive Omnitool's per-permanent pump over the Aura controller's creatures
// (the host counts itself), and Champion's Helm's conditional keyword read on
// every recompute, so a creature that becomes a Human gains first strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "553b9606-0519-4b85-b6ae-acc126e537e1",
		Name:         "Hope Against Hope",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, effects.OldB289CreaturesYouControl),
			effects.GrantToAttachedWhile(effects.OldB289HostIsHuman, "first strike"),
		},
	})
}
