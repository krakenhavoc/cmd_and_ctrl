package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chill Haunting — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, exile X creature cards
//	 from your graveyard.
//	 Target creature gets -X/-X until end of turn."
//
// The X form of the graveyard-exile cast cost (#1948, ADR 0189's
// 2026-10-16 amendment), narrowed to creature cards: X is announced
// with the cast (CR 601.2b), bounded by the creature cards the graveyard
// holds (CR 118.3), and is both the number of cards named on exile_ids
// and the size of the -X/-X (CR 107.3i). The cards are read as they are
// in the graveyard (printed types).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ebe61471-e236-4ff8-9416-151970b5efd6",
		Name:           "Chill Haunting",
		Purpose:        effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.ExileXFromGraveyardCost("X creature cards", func(c game.Card) bool { return c.IsCreature() }),
		Targets:        effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			targets := ctx.LegalTargets()
			if x <= 0 || len(targets) == 0 {
				return nil
			}
			return effects.BoostUntilEOT{Target: targets[0].ID, Power: -x, Toughness: -x,
				Label: "Chill Haunting — -X/-X until end of turn"}.Apply(ctx)
		},
	})
}
