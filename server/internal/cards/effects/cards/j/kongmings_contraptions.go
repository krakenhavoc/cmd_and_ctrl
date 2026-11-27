package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kongming's Contraptions — Creature — Human Soldier {3}{W}, 2/4 (Portal Three
// Kingdoms, S82 #5622):
//
//	"{T}: This creature deals 2 damage to target attacking creature. Activate only
//	 during the declare attackers step and only if you've been attacked this step."
//
// Deep Wood's window as an activation condition (OldB708YouWereAttackedThisStep):
// the declare attackers step, with a creature declared attacking you (not a
// planeswalker you control); a creature removed from combat since still counts.
// The target is an attacking creature, chosen as the ability is activated and
// re-judged as it resolves; the damage is dealt by the Contraptions, as it last
// existed if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3327ffe-f47f-4f5b-b929-a7706ef4932d",
		Name:         "Kongming's Contraptions",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: This creature deals 2 damage to target attacking creature. Activate only during the declare attackers step and only if you've been attacked this step.",
			Purpose:   effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:      effects.TapCost(),
			Condition: effects.OldB708YouWereAttackedThisStep(),
			Targets:   effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect:    effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
