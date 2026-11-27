package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunscape Apprentice — Creature — Human Wizard {W}, 1/1 (Invasion, S82 OLD
// #5616):
//
//	"{G}, {T}: Target creature gets +1/+1 until end of turn.
//	 {U}, {T}: Put target creature you control on top of its owner's library."
//
// Two tap abilities, so summoning sickness bars both (CR 302.6). The first
// pumps any creature; the second tucks one of the controller's own creatures
// (the Apprentice itself included), each target judged again as the ability
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ab1411e-b255-4a6e-b81a-5de7b2d4ea12",
		Name:         "Sunscape Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}, {T}: Target creature gets +1/+1 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
				Cost:    game.AbilityCost{Mana: "{G}", Tap: true},
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Sunscape Apprentice — +1/+1 until end of turn", 1, 1),
			},
			{
				Label:   "{U}, {T}: Put target creature you control on top of its owner's library.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    game.AbilityCost{Mana: "{U}", Tap: true},
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.OldB651PutTheTargetOnTopOfItsOwnersLibrary,
			},
		},
	})
}
