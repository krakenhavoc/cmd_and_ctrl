package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imposing Visage — Enchantment — Aura {R} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature has menace."
//
// A layer 6 keyword grant on the attached creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f67058b-ed14-4e3c-9af3-d61570870e36",
		Name:         "Imposing Visage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("menace")},
	})
}
