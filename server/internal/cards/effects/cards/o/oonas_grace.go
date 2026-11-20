package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oona's Grace — Instant {2}{U}:
//
//	"Target player draws a card.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "380e9992-b5d2-4fbe-a8c3-c37220846e0a",
		Name:             "Oona's Grace",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Retrace("{2}{U}")},
		Targets:          effects.TargetPlayer("target player"),
		Purpose:          effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return effects.DrawCards{Player: item.Targets[0].ID, N: 1}.Apply(ctx)
		},
	})
}
