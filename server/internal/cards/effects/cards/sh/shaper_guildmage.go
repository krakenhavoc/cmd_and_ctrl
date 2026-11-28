package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shaper Guildmage — Creature — Human Wizard {U}, 1/1 (Mirage, S82 OLD b797,
// #5633):
//
//	"{W}, {T}: Target creature gains first strike until end of turn.
//	 {B}, {T}: Target creature gets +1/+0 until end of turn."
//
// Two tap abilities on their own target creature (Armorer Guildmage's
// shape): a keyword grant and a boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfbce321-d3c1-4b76-a3fa-27f875015085",
		Name:         "Shaper Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}: Target creature gains first strike until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NeoB07TargetGetsUntilEOT("Shaper Guildmage — first strike", 0, 0, "first strike"),
			},
			{
				Label:   "{B}, {T}: Target creature gets +1/+0 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NeoB07TargetGetsUntilEOT("Shaper Guildmage", 1, 0),
			},
		},
	})
}
