package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonscale Boon — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Instant {3}{G}: Put two +1/+1 counters on target creature and untap it.
//
// Soldier On's body, shared (OldB346TwoCountersAndUntapTheTarget).

func init() {
	effects.Register(effects.Spec{
		OracleID:     "629b3aee-b8fd-495a-a988-815a0aaf3e82",
		Name:         "Dragonscale Boon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.OldB346TwoCountersAndUntapTheTarget,
	})
}
