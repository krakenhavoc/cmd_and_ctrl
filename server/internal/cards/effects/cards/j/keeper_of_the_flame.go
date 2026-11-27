package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of the Flame — Creature — Human Wizard {R}{R}, 1/2 (Exodus, S82 OLD
// b738, #5626):
//
//	"{R}, {T}: Choose target opponent who has more life than you do as you
//	 activate this ability. This creature deals 2 damage to that player."
//
// The life comparison is read only as the ability is announced (CR 602.2b):
// the target clause carries MoreLifeAsActivated, which the activation gate and
// the legal set read and the CR 608.2b re-check does not, so the damage is
// still dealt if the totals change in response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cf12b01-7c0a-488f-9527-404a545c70c4",
		Name:         "Keeper of the Flame",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: Choose target opponent who has more life than you do as you activate this ability. This creature deals 2 damage to that player.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.WithMoreLifeAsActivated(effects.TargetPlayer("target opponent", effects.Opponent())),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
