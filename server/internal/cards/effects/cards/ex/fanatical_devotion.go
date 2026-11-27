package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fanatical Devotion — Enchantment {2}{W} (Nemesis, S82 OLD batch 679, #5618):
//
//	"Sacrifice a creature: Regenerate target creature."
//
// Goblin Chirurgeon's regeneration with any creature as the cost. The
// sacrifice is paid as the ability is activated (CR 602.2b), so a creature
// that sacrifices itself to regenerate itself regenerates nothing (CR
// 701.19b: it is no longer on the battlefield), and sacrifice is never
// destruction, so the creature eaten is not saved by a shield of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34e56412-4649-481d-be51-357b331d0b75",
		Name:         "Fanatical Devotion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
