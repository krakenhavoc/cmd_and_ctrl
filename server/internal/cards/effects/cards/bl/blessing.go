package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessing — Enchantment — Aura {W}{W} (Magic 2014, S82 old sets):
//
//	"Enchant creature
//	 {W}: Enchanted creature gets +1/+1 until end of turn."
//
// Flowstone Blade's shape: the Aura's own activated ability (its controller
// pays, whoever controls the creature), boosting the creature the Aura is
// attached to as it resolves, and nothing happens if it is no longer attached
// (CR 608.2b). Each activation is its own boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c84d8da-2bfb-4618-89a0-7d9ed604e854",
		Name:         "Blessing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Enchanted creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.S82OldB399PumpEnchanted("Blessing", 1, 1),
		}},
	})
}
