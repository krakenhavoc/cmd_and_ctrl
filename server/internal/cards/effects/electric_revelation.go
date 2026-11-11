package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "c92af0f4-e5f1-4731-a781-b2ca360867a7",
		Name:             "Electric Revelation",
		Completeness:     CompletenessFull,
		AdditionalCost:   DiscardCost(1),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
