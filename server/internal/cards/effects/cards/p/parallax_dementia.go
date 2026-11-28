package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parallax Dementia — Enchantment — Aura {1}{B} (Nemesis, S82 OLD, #5618):
//
//	"Enchant creature
//	 Fading 1
//	 Enchanted creature gets +3/+2.
//	 When this Aura leaves the battlefield, destroy enchanted creature. That
//	 creature can't be regenerated."
//
// Fading is the shared pair, the pump is PumpAttached, and the leave trigger
// destroys the creature the Aura was attached to as it last existed (CR 608.2h),
// ignoring regeneration (CR 701.19c). An Aura that left because its creature
// did destroys nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ec7a7fb-0bf9-47fa-af2b-5b13b3af3947",
		Name:         "Parallax Dementia",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Replacements: []game.ReplacementEffect{effects.Fading(1)},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 2)},
		Triggered: []game.TriggeredAbility{
			effects.FadingUpkeep(),
			effects.WhenThisLeaves("Parallax Dementia — destroy enchanted creature, it can't be regenerated", effects.OldB678DestroyEnchantedOnLeave),
		},
	})
}
