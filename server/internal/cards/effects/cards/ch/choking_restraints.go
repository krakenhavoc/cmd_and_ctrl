package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Choking Restraints — Enchantment — Aura {2}{W} (Eldritch Moon, #5541):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {3}{W}{W}, Sacrifice this Aura: Exile enchanted creature."
//
// Spiral into Solitude's shape without the blight: the restriction on
// the host, and an activation whose sacrifice is a cost, so the exile
// reads the creature the Aura was attached to as it last existed
// (exileEnchantedCreature).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42bc33dd-3bba-462d-b272-2e8bd58c649e",
		Name:         "Choking Restraints",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}{W}{W}, Sacrifice this Aura: Exile enchanted creature.",
			Cost:  effects.Plus(effects.ManaCost("{3}{W}{W}"), effects.SacrificeThis()),
			// Removes a permanent with no target (ADR 0142's AnswerRemove).
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.ExileEnchantedCreature,
		}},
	})
}
