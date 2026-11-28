package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snow Fortress — Artifact Creature — Wall {5}, 0/4 (Ice Age, S82 OLD batch 840,
// #5636):
//
//	"Defender
//	 {1}: This creature gets +1/+0 until end of turn.
//	 {1}: This creature gets +0/+1 until end of turn.
//	 {3}: This creature deals 1 damage to target creature without flying that's
//	 attacking you."
//
// Two layer 7c boosts that end at cleanup (CR 514.2) and Balduvian Trading Post's
// damage to an attacker: the target must be attacking its controller (the player
// who activates it), judged as the ability is announced and again as it resolves.
// Flying is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7709ed84-ed71-4dd2-b002-b257418dd79e",
		Name:            "Snow Fortress",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{
			effects.SelfPumpRow("Snow Fortress", "{1}", 1, 0),
			effects.SelfPumpRow("Snow Fortress", "{1}", 0, 1),
			{
				Label:   "{3}: This creature deals 1 damage to target creature without flying that's attacking you.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.ManaCost("{3}"),
				Targets: effects.TargetCreature("target creature without flying that's attacking you", effects.AttackingYou(), effects.Not(effects.HasKeyword("flying"))),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			},
		},
	})
}
