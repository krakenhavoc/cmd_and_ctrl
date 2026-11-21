package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Laughing Mad — Instant {2}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards.
//	 Flashback {3}{R} (You may cast this card from your graveyard for
//	 its flashback cost and any additional costs. Then exile it.)"
//
// The additional cost is paid on EITHER cast path — Spec.AdditionalCost
// is orthogonal to Spec.AlternativeCosts, so the printed "and any
// additional costs" needs no extra plumbing: a flashback cast already
// pays it the same way a hand cast does, with the card already on the
// stack (Thrill of Possibility's reasoning, one zone over).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4f65e6a9-0d90-44b9-9b76-00814db2dbd8",
		Name:             "Laughing Mad",
		Completeness:     effects.CompletenessFull,
		AdditionalCost:   effects.DiscardCost(1),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
