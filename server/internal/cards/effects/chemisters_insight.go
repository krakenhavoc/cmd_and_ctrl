package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chemister's Insight — Instant {3}{U} (Ravnica: Clue Edition, #5929):
//
//	"Draw two cards.
//	 Jump-start (You may cast this card from your graveyard by discarding a
//	 card in addition to paying its other costs. Then exile this card.)"
//
// Radical Idea's body at a different cost and draw count: jump-start is
// cmdr2S39JumpStart at the card's own mana cost (CR 702.133a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a2cb30d7-c0a9-4f79-bc19-fc88c4986725",
		Name:             "Chemister's Insight",
		Purpose:          game.Purpose{Draws: 2},
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{cmdr2S39JumpStart("{3}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
