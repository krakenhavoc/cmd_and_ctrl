package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radical Idea — Instant {1}{U}:
//
//	"Draw a card.
//	 Jump-start (You may cast this card from your graveyard by discarding a
//	 card in addition to paying its other costs. Then exile this card.)"
//
// The only jump-start card: cmdr2S39JumpStart is flashback at the card's own
// cost with a discard from hand added (CR 702.133a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "fd51dce2-8e8a-4686-910e-1bbc2825d190",
		Name:             "Radical Idea",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{cmdr2S39JumpStart("{1}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
