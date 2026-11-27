package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nightscape Apprentice (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Zombie Wizard {B}, 1/1:
//	"{U}, {T}: Put target creature you control on top of its owner's library.
//	{R}, {T}: Target creature gains first strike until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "429de3c3-3f2e-4d55-a04a-f414912aff3a",
		Name:         "Nightscape Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, {T}: Put target creature you control on top of its owner's library.",
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Effect:  effects.S82OldB645TargetToTopOfLibrary,
			},
			effects.S82OldB645TargetGainsRow(effects.Plus(effects.ManaCost("{R}"), effects.TapCost()), "{R}, {T}", "creature", "first strike"),
		},
	})
}
