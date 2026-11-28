package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Guildmage — Creature — Human Wizard {B}, 1/1 (Mirage, #5633):
//
//	"{U}, {T}: Put target creature you control on top of its owner's library.
//	 {R}, {T}: This creature deals 1 damage to any target and 1 damage to you."
//
// The second ability's two damages are one simultaneous event, and an illegal
// target at resolution does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdda63d7-e3ea-455c-b7d4-c6cfdda2703c",
		Name:         "Shadow Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, {T}: Put target creature you control on top of its owner's library.",
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.PutChosenTargetOnTopOfLibrary,
			},
			{
				Label:   "{R}, {T}: This creature deals 1 damage to any target and 1 damage to you.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.OldB795DealsOneToTargetAndToYou,
			},
		},
	})
}
