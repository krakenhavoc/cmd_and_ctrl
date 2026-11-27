package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keldon Mantle — Enchantment — Aura {1}{R} (Planeshift, #5615):
//
//	"Enchant creature
//	 {B}: Regenerate enchanted creature.
//	 {R}: Enchanted creature gets +1/+0 until end of turn.
//	 {G}: Enchanted creature gains trample until end of turn."
//
// Three abilities of the Aura itself, each finding its host as it resolves
// (Blessing of Leeches' regeneration, Shiv's Embrace's firebreathing and
// Pemmin's Aura's keyword grant), so an Aura that has fallen off by then does
// nothing. Anyone who controls the Aura may pay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "170cc0c4-0fba-4751-8613-e934b0438364",
		Name:         "Keldon Mantle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: Regenerate enchanted creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.RegenerateAttachedHost,
			},
			{
				Label:   "{R}: Enchanted creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.S82OldB399PumpEnchanted("Keldon Mantle", 1, 0),
			},
			{
				Label:   "{G}: Enchanted creature gains trample until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{G}"),
				Effect:  effects.ScgB1HostGainsUntilEOT("trample", "Keldon Mantle — trample until end of turn"),
			},
		},
	})
}
