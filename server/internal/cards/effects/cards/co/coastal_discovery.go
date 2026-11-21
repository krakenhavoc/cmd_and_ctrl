package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coastal Discovery — Sorcery {3}{U}:
//
//	"Draw two cards.
//	 Awaken 4—{5}{U}"
//
// ADR 0135 §3 (#2411): draw two, then the awaken land (CR 702.113a). The
// land is the spell's only target, so an awaken cast whose land is gone
// does not resolve and draws nothing (CR 608.2b; the card's ruling).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7c84690-8c7c-41a6-b430-a9771030f903",
		Name:         "Coastal Discovery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		AlternativeCosts: []game.AlternativeCost{
			effects.Awaken(4, "{5}{U}", nil),
		},
		OnResolve: effects.AwakenAfter(4, func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		}),
	})
}
