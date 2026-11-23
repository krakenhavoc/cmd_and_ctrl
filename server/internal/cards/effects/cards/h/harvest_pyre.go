package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvest Pyre — Instant {1}{R}:
//
//	"As an additional cost to cast this spell, exile X cards from your
//	 graveyard.
//	 Harvest Pyre deals X damage to target creature."
//
// The X form of the graveyard-exile cast cost (#1948, ADR 0189's
// 2026-10-16 amendment). The mana cost has no {X}, so X lives only in
// the additional cost: it is announced with the cast (CR 601.2b, the
// client opens its X prompt because the cost demands one), bounded by
// the cards the graveyard holds (CR 118.3), and it is both the number of
// cards named on exile_ids and the damage (CR 107.3i). A target that
// left in response gets nothing (CR 608.2b); the cards stay exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "00e36265-0cd8-4084-adf5-d1ed44f0d689",
		Name:           "Harvest Pyre",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.ExileXFromGraveyardCost("X cards", nil),
		Targets:        effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealXToTarget(item, ctx)
		},
	})
}
