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
// Retrace (CR 702.34) needs a cast path from the graveyard bound to
// a cost, the same shape Flashback and Escape use (Spec.CastableZones
// + Spec.AlternativeCosts) — but the price itself is "discard a
// land," and every discard-cost constructor the catalog has
// (DiscardCost) takes a card count with no type filter. Retrace is
// the only card in this batch that prints it, so per AGENTS.md §7
// this is noted here rather than in docs/engine-seams.md. Audited
// under #1112; not modelled.
func init() {
	Register(Spec{
		OracleID:     "13e94530-defb-4c9b-9ec7-bb7789ec2630",
		Name:         "Decaying Time Loop",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Retrace isn't implemented — this spell can only be cast from your hand, never later from the graveyard by discarding a land.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n, err := discardWholeHand(ctx.Game, item.Controller)
			if err != nil || n == 0 {
				return err
			}
			return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
