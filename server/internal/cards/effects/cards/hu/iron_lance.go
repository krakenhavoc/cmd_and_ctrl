package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iron Lance — Artifact {2} (Mercadian Masques, S82 OLD b681, #5619):
//
//	"{3}, {T}: Target creature gains first strike until end of turn."
//
// Boros Guildmage's first-strike grant on a mana-and-tap cost, any
// controller's creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09e588c2-0fb8-4c30-aff3-433db7a07b6e",
		Name:         "Iron Lance",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("first strike", "Iron Lance — first strike until end of turn"),
		}},
	})
}
