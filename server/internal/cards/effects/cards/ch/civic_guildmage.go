package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Civic Guildmage — Creature — Human Wizard {W}, 1/1 (Mirage, #5633):
//
//	"{G}, {T}: Target creature gets +0/+1 until end of turn.
//	 {U}, {T}: Put target creature you control on top of its owner's library."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e93cb9a-44c8-49fb-b6a7-bfdf7bd725fe",
		Name:         "Civic Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}, {T}: Target creature gets +0/+1 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 1)),
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Civic Guildmage — +0/+1", 0, 1),
			},
			{
				Label:   "{U}, {T}: Put target creature you control on top of its owner's library.",
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.PutChosenTargetOnTopOfLibrary,
			},
		},
	})
}
