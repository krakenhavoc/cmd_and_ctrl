package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Kiss — Enchantment — Aura {2}{B} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 {B}, Pay 1 life: Enchanted creature gets +2/+2 until end of turn. Activate
//	 no more than three times each turn."
//
// Fire-Belly Changeling's per-turn tally, with the life paid as part of the cost.
//
// No simplification.
func init() {
	const label = "{B}, Pay 1 life: Enchanted creature gets +2/+2 until end of turn. Activate no more than three times each turn."
	effects.Register(effects.Spec{
		OracleID:     "8ea941aa-ee0c-4830-912a-0c672546aa0c",
		Name:         "Soul Kiss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.Plus(effects.ManaCost("{B}"), effects.PayLife(1)),
			Condition: effects.AtMostNActivationsPerTurn(label, 3),
			Effect:    effects.S82OldB399PumpEnchanted("Soul Kiss", 2, 2),
		}},
	})
}
