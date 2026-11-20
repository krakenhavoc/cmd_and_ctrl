package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warning — Instant {W}:
//
//	"Prevent all combat damage that would be dealt by target attacking creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted attacker, pinned as the spell resolves (CR 400.7).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5ba0f0f-65c5-4ffa-987a-f320b401ec8f",
		Name:         "Warning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:    effects.ShieldAgainstTargetsSpell(true),
	})
}
