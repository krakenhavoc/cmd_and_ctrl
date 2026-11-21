package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prison Sentence — Enchantment — Aura {2}{W} (The Brothers' War):
//
//	"Enchant creature
//	 When this Aura enters, scry 2.
//	 Enchanted creature can't attack or block, and its activated abilities
//	 can't be activated."
//
// Arrest's restriction with Angelic Gift's enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d9f2fe0-5657-45b4-9431-36e0f6c8eba2",
		Name:         "Prison Sentence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm, Scry: 2},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Prison Sentence — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
