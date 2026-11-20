package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildheart Invoker — Creature — Elf Shaman {2}{G}{G}, 4/3
// (Foundations, #4465):
//
//	"{8}: Target creature gets +5/+5 and gains trample until end of
//	 turn."
//
// Dr. Madison Li's pump-and-grant row with a mana cost: the target gets
// +5/+5 and trample if it is still legal as the ability resolves (CR
// 608.2b). Every use has a target, so the row declares no Answers
// (ADR 0142).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff8fb796-af5d-4d64-ab3c-7f2469dfa6bf",
		Name:         "Wildheart Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Target creature gets +5/+5 and gains trample until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 5, 5, "trample")),
			Cost:    effects.ManaCost("{8}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PumpAndGrantFirstTarget(5, 5, "Wildheart Invoker — +5/+5 and trample", "trample"),
		}},
	})
}
