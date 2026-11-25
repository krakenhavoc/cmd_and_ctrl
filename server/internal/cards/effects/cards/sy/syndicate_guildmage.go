package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syndicate Guildmage — Creature — Human Cleric {W}{B}, 2/2 (Ravnica Allegiance):
//
//	"{1}{W}, {T}: Tap target creature with power 4 or greater.
//	 {4}{B}, {T}: This creature deals 2 damage to target opponent or planeswalker."
//
// The tap target is legal only while its power, read through the layered view, is
// 4 or more (CR 608.2b re-checks it on resolution).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec484fd1-aa9a-4056-becf-1d8c710272fd",
		Name:         "Syndicate Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}, {T}: Tap target creature with power 4 or greater.",
				Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
				Effect:  effects.TapTheFirstLegalTarget,
			},
			{
				Label:   "{4}{B}, {T}: This creature deals 2 damage to target opponent or planeswalker.",
				Cost:    effects.Plus(effects.ManaCost("{4}{B}"), effects.TapCost()),
				Targets: effects.TargetOpponentOrPlaneswalker(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.DealDamageToTheTarget(2),
			},
		},
	})
}
