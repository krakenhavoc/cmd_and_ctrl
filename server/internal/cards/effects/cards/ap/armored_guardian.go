package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armored Guardian (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Cat Soldier {3}{W}{U}, 2/5:
//	"{1}{W}{W}: Target creature you control gains protection from the color of your choice until end of turn.
//	{1}{U}{U}: This creature gains shroud until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46752289-01e8-4839-8889-49db1efd2e1a",
		Name:         "Armored Guardian",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}{W}: Target creature you control gains protection from the color of your choice until end of turn.",
				Cost:    effects.ManaCost("{1}{W}{W}"),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
				Effect:  effects.S82OldB645ProtectionFromChosenColor("Armored Guardian"),
			},
			effects.ConB1ThisGainsRow("{1}{U}{U}", "shroud", "Armored Guardian", game.AnswerProtect),
		},
	})
}
