package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eviscerator's Insight — Instant {1}{B} (Modern Horizons 3, #5948):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Draw two cards.
//	 Flashback {4}{B} (You may cast this card from your graveyard for its
//	 flashback cost and any additional costs. Then exile it.)"
//
// Altar's Reap's body with the shared Flashback: the sacrifice is an
// additional cost, so a flashback cast pays it too (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "81a398ee-e534-43b7-88fe-45155d6aba3a",
		Name:             "Eviscerator's Insight",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 2},
		AdditionalCost:   effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
