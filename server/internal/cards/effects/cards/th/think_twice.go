package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Think Twice — Instant {1}{U}:
//
//	"Draw a card.
//	 Flashback {2}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The reference flashback cantrip: one card now, one more later out of
// the graveyard for a little more mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "fa85c5a2-8e83-4624-a35a-a0bbf17ecbb4",
		Name:             "Think Twice",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
