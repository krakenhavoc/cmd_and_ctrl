package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// One Thousand Lashes — Enchantment — Aura {2}{W}{B} (Gatecrash, S82 OLD b407,
// #5565):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block, and its activated abilities
//	 can't be activated.
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player loses 1 life."
//
// Arrest's restriction and Stab Wound's upkeep trigger: the loser is the
// controller of the enchanted creature as the upkeep begins, not the Aura's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3b259e9-f9c7-4830-91fa-6e3d9f4cc1e3",
		Name:         "One Thousand Lashes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"One Thousand Lashes — that player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -1)
				}),
		},
	})
}
