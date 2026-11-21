package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feed the Cycle — Instant {1}{B} (Bloomburrow, #4467):
//
//	"As an additional cost to cast this spell, forage or pay {B}. (To
//	 forage, exile three cards from your graveyard or sacrifice a Food.)
//	 Destroy target creature or planeswalker."
//
// The either/or additional cost (ADR 0100 §2) with forage as one side
// (#5184, ForageOrPay): the caster picks one of three branches — exile
// three cards from their graveyard, sacrifice a Food, or pay {B}. The
// forage branches are greyed when the graveyard holds fewer than three
// cards or there is no Food (CR 118.3), and paying either one is a
// forage (CR 701.61a), so "whenever you forage" sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "c338eedb-9b70-4b77-b5d6-8ea74b7bea3e",
		Name:           "Feed the Cycle",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.ForageOrPay("{B}"),
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve:      effects.DestroyTheTargetPermanent,
	})
}
