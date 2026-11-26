package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fateful Showdown — Instant {2}{R}{R} (Kaladesh, #5539):
//
//	"Fateful Showdown deals damage to any target equal to the number of
//	 cards in your hand. Discard all the cards in your hand, then draw that
//	 many cards."
//
// The damage is the hand as the spell resolves (it has left the hand by
// then), read once. Then Tolarian Winds' body: the hand is discarded whole,
// and the draw is the count of what left, read before the draw so the new
// cards are not counted. With the only target illegal the spell does
// nothing, so the hand stays (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "729b7ce8-7e7f-4db9-b3ce-67aba46a7b0d",
		Name:         "Fateful Showdown",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SourceDealsComputedDamageToEachLegalTarget(func(c *effects.Context) int {
					if p := c.PlayerByID(c.Controller()); p != nil {
						return p.Hand.Size()
					}
					return 0
				})(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				n, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
				ctx.RecordStep(fatefulShowdownRecord, nil, n)
				return err
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: ctx.StepRecord(fatefulShowdownRecord).N()}.Apply(ctx)
			}),
		),
	})
}

// fatefulShowdownRecord is the number of cards discarded.
const fatefulShowdownRecord = "fateful-showdown/discarded"
