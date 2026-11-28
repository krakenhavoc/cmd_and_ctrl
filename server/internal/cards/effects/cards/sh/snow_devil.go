package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Snow Devil — Enchantment — Aura {1}{U} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 Enchanted creature has first strike as long as it's blocking and you
//	 control a snow land."
//
// A layer 6 flying grant, and a conditional first strike read on every
// recompute: the creature is blocking and the Aura's controller has a snow land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e743b4ec-65bf-4e8d-ad87-6746a2e6ba73",
		Name:         "Snow Devil",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying"),
			effects.SnowDevilFirstStrikeB830(),
		},
	})
}
