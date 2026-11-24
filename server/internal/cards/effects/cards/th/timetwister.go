package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timetwister — Sorcery {2}{U}: "Each player shuffles their hand and
// graveyard into their library, then draws seven cards. (Then put
// Timetwister into its owner's graveyard.)" (Alpha, #5948). One of the
// Power Nine Oracle of the Alpha conjures (ADR 0099 amendment
// 2026-11-24). Echo of Eons' resolution; the reminder is what any
// resolved sorcery does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c823e687-6311-4c99-974b-fd77d204141a",
		Name:         "Timetwister",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.EchoOfEonsResolve,
	})
}
