package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decaying Time Loop — Instant for {3}{R}:
//
//	"Discard all the cards in your hand, then draw that many cards.
//	 Retrace"
//
// A one-sided Windfall at instant speed. The count is taken before
// the draw, so an empty hand draws nothing — which is the whole
// reason to cast it in response to something rather than on your own
// turn with three cards left.
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: CastableZones opens the graveyard and
// Retrace(cost) prices it (ADR 0066, 2026-10-07 amendment, #2528). The
// spell goes back to the graveyard when it resolves, so it can be retraced
// again, at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "13e94530-defb-4c9b-9ec7-bb7789ec2630",
		Name:             "Decaying Time Loop",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n, err := discardWholeHand(ctx.Game, item.Controller)
			if err != nil || n == 0 {
				return err
			}
			return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
