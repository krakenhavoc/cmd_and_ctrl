package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Electric Revelation — Instant {2}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards.
//	 Flashback {3}{R} (You may cast this card from your graveyard for its
//	 flashback cost and any additional costs. Then exile it.)"
//
// Laughing Mad's card: the discard is paid as the spell is cast, from the
// graveyard too (flashback keeps the additional cost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c92af0f4-e5f1-4731-a781-b2ca360867a7",
		Name:             "Electric Revelation",
		Completeness:     effects.CompletenessFull,
		AdditionalCost:   effects.DiscardCost(1),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
