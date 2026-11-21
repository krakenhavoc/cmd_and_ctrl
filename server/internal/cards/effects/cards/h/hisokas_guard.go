package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hisoka's Guard — Creature — Human Wizard (1/1) for {1}{U}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {1}{U}, {T}: Target creature you control other than this creature has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)"
//
// ADR 0109 owner decision 4: a continuous effect from a resolving
// ability (CR 611.2), timed by "for as long as this creature remains
// tapped" (CR 611.2b). The effect is a ScopedEffect record, data a
// restore point carries; it ends the moment the creature untaps, and
// for good. "You may choose not to untap" is the untap opt-out, which
// is how its controller keeps the effect going.
//
// "Other than this creature" is the clause's ExcludeSource
// (effects.Another).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6136f019-da8f-471e-9ac5-41e535aa8a21",
		Name:         "Hisoka's Guard",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Hisoka's Guard — you may choose not to untap this creature")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Target creature you control other than this creature has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("target creature you control other than this creature", effects.YouControl())),
			Effect:  effects.TargetGetsWhileThisRemainsTapped("Hisoka's Guard — shroud while tapped", game.AddKeywordsMod("shroud")),
		}},
	})
}
