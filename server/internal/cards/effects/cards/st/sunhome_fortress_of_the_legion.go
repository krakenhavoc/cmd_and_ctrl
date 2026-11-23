package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunhome, Fortress of the Legion — Land:
//
//	"{T}: Add {C}.
//	 {2}{R}{W}, {T}: Target creature gains double strike until end of
//	 turn."
//
// Axgard Cavalry's grant on a land: the keyword is granted to the
// target as the ability resolves and lasts until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a9f8344c-1705-4254-81d6-aa05e0c69c29",
		Name:          "Sunhome, Fortress of the Legion",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}{W}, {T}: Target creature gains double strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")),
			Cost:    effects.Plus(effects.ManaCost("{2}{R}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("double strike", "Sunhome, Fortress of the Legion — double strike until end of turn"),
		}},
	})
}
