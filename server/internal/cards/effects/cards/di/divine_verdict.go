package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divine Verdict — Instant {3}{W} (Magic 2010, #5585):
//
//	"Destroy target attacking or blocking creature."
//
// Sudden Strike's target clause: only a creature that is attacking or
// blocking can be chosen, and it is checked again as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3070f997-24f2-471d-af96-a2f3b588198e",
		Name:         "Divine Verdict",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
