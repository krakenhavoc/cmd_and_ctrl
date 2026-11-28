package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fylgja — Enchantment — Aura {W} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 This Aura enters with four healing counters on it.
//	 Remove a healing counter from this Aura: Prevent the next 1 damage that
//	 would be dealt to enchanted creature this turn.
//	 {2}{W}: Put a healing counter on this Aura."
//
// Ursine Fylgja's healing counters on the Aura, with the shield on the creature
// the Aura enchants as the ability resolves.
//
// No simplification.
func init() {
	const healing = "healing"
	effects.Register(effects.Spec{
		OracleID:     "df0e5dd8-65e0-474c-ac32-6e0f6f16b8e6",
		Name:         "Fylgja",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(healing, 4, "Enters with four healing counters on it.")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove a healing counter from this Aura: Prevent the next 1 damage that would be dealt to enchanted creature this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.RemoveCountersFromThis(healing, 1),
				Effect:  effects.PreventNextOneOnEnchantedB830,
			},
			{
				Label:   "{2}{W}: Put a healing counter on this Aura.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{2}{W}"),
				Effect:  effects.Do(effects.CounterOnThis{Kind: healing, N: 1}),
			},
		},
	})
}
