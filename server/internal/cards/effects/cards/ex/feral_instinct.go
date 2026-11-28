package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Instinct — Instant {1}{G} (Visions, S82 OLD batch 787, #5631):
//
//	"Target creature gets +1/+1 until end of turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// A layer 7c boost (CR 613.4c) that ends at cleanup (CR 514.2), then the
// delayed draw at the next upkeep at the table, whoever's turn it is
// (CR 603.7a, next_upkeep_draw.go). With its one target gone the spell does
// not resolve at all, so nothing is drawn either (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8f3fc11-25b4-4a23-9d79-f380a14f0420",
		Name:         "Feral Instinct",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTheTargetUntilEOT(1, 1, "Feral Instinct — +1/+1 until end of turn")(item, ctx); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Feral Instinct — draw a card").Apply(ctx)
		},
	})
}
