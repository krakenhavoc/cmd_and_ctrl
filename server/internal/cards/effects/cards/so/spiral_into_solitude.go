package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiral into Solitude — Enchantment — Aura {1}{W}:
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {1}{W}, Blight 1, Sacrifice this Aura: Exile enchanted creature. (To
//	 blight 1, put a -1/-1 counter on a creature you control.)"
//
// Pacifism, with Redemption Arc's exile behind a cost that blights
// (BlightN, #4684) and sacrifices the Aura. The counter goes on before
// the Aura is sacrificed, and "enchanted creature" is the creature the
// Aura was attached to as it left (CR 608.2h), so the exile still finds
// it. It does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8599f06-f030-4575-8497-307a4d7ec182",
		Name:         "Spiral into Solitude",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{W}, Blight 1, Sacrifice this Aura: Exile enchanted creature.",
			Cost:  effects.Plus(effects.ManaCost("{1}{W}"), effects.BlightN(1), effects.SacrificeThis()),
			// Removes a permanent with no target (ADR 0142's AnswerRemove).
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.ExileEnchantedCreature,
		}},
	})
}
