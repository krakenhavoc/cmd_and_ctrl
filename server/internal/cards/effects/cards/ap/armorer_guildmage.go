package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armorer Guildmage — Creature — Human Wizard {R}, 1/1 (Mirage, S82 OLD b797,
// #5633):
//
//	"{B}, {T}: Target creature gets +1/+0 until end of turn.
//	 {G}, {T}: Target creature gets +0/+1 until end of turn."
//
// Two tap abilities, each on its own target creature, through the shared
// targeted pump body (Advance Scout).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b75b74c-6003-4b34-bae3-2c9e2f880042",
		Name:         "Armorer Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}, {T}: Target creature gets +1/+0 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NeoB07TargetGetsUntilEOT("Armorer Guildmage", 1, 0),
			},
			{
				Label:   "{G}, {T}: Target creature gets +0/+1 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 1)),
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NeoB07TargetGetsUntilEOT("Armorer Guildmage", 0, 1),
			},
		},
	})
}
