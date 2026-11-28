package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suq'Ata Firewalker — Creature — Human Wizard {1}{U}{U}, 0/1 (Mirage, #5633):
//
//	"This creature can't be the target of red spells or abilities from red
//	 sources.
//	 {T}: This creature deals 1 damage to any target."
//
// Only a targeting bar, not protection: red damage and blocks are untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9c0bb7e-8dfc-481c-80f7-b6652b63df39",
		Name:         "Suq'Ata Firewalker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CantBeTargetedByColor("R", "This creature can't be the target of red spells or abilities from red sources."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
