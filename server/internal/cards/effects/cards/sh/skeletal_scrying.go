package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeletal Scrying — Instant {X}{B}:
//
//	"As an additional cost to cast this spell, exile X cards from your
//	 graveyard.
//	 You draw X cards and you lose X life."
//
// The X form of the graveyard-exile cast cost (#1948, ADR 0189's
// 2026-10-16 amendment): one X is announced with the cast (CR 601.2b),
// and it is the {X} in the mana cost, the number of cards named on
// exile_ids and the number the spell draws and loses (CR 107.3i). The
// caster can't announce more than the graveyard holds (CR 118.3). The
// cards are exiled with the spell on the stack, so a countered Scrying
// keeps them in exile. Cast without paying its mana cost, X is 0
// (CR 107.3b) and nothing is exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "561c68d8-b727-48d8-a832-2e42d1345fac",
		Name:           "Skeletal Scrying",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.ExileXFromGraveyardCost("X cards", nil),
		// The life loss waits for a prompt a draw asks (a dredge offer; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(effects.DrawXStep(), effects.LoseXStep()),
	})
}
