package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Transfer — {4}{U} Sorcery (Starter 1999, S82 OLD batch 702, #5620):
//
//	"If the difference between your life total and target player's life total
//	 is 5 or less, exchange life totals with that player."
//
// Judged as it resolves (CR 608.2c) from the two life totals (S82OldB702PsychicTransfer);
// the exchange is Profane Transfusion's (ExchangeLifeTotalsThen).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5554d422-6513-4c7e-866d-9d5f0403fc80",
		Name:         "Psychic Transfer",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.S82OldB702PsychicTransfer,
	})
}
