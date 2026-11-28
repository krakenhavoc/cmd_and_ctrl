package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra Paladin — Creature — Human Knight {2}{W}{W}, 2/2 (Homelands, S82 OLD
// b821, #5635):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this turn.
//	{1}{W}{W}, {T}: Target creature gains vigilance until end of turn."
//
// Benevolent Ancestor's prevention shield and Helm of Chatzuk's keyword grant,
// each behind its own cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fdbbca7-d1d0-4834-84df-2c2d9953f83a",
		Name:         "Serra Paladin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextOneDamageToTheTarget("Serra Paladin"),
			},
			{
				Label:   "{1}{W}{W}, {T}: Target creature gains vigilance until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance")),
				Cost:    effects.Plus(effects.ManaCost("{1}{W}{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.EbTargetCreatureGainsUntilEOT("vigilance", "Serra Paladin — vigilance until end of turn"),
			},
		},
	})
}
