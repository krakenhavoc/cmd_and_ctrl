package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuna's Whistle — Instant {1}{G}{G}:
//
//	"Reveal cards from the top of your library until you reveal a
//	 creature card. Put that card into your hand and the rest on the
//	 bottom of your library in a random order. When you reveal a
//	 creature card this way, put X +1/+1 counters on target creature you
//	 control, where X is the mana value of that card."
//
// Artificer Class's reveal-until-to-hand with a creature card, plus a
// CR 603.12 reflexive trigger created only when a creature card was
// revealed (a library with none reveals itself and creates nothing).
// The trigger's target is chosen as it goes on the stack, after the
// spell has finished, and X is the revealed card's mana value, carried
// on the trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39bf633c-41a6-496d-8aef-1ab7befb503d",
		Name:         "Yuna's Whistle",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
				"Yuna's Whistle — reveal until a creature card")
			if hit != uuid.Nil {
				mv := 0
				if c, ok := ctx.Game.LookupCardForEffect(hit); ok {
					mv = c.ManaValue()
				}
				if err := (ReflexiveTrigger{
					Label:  "Yuna's Whistle — put X +1/+1 counters on target creature you control",
					Body:   yunasWhistleCountersBody,
					Params: game.EffectParams{Amount: mv},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  run,
				Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
				All:    true,
				Label:  "Yuna's Whistle — put the creature card into your hand",
				Then:   TakeRestOnBottomInRandomOrder,
			}.Apply(ctx)
		},
	})
}

var yunasWhistleCountersBody = game.ReflexiveBody("yunas-whistle/counters",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		ts := ctx.LegalTargets()
		if len(ts) == 0 || p.Amount <= 0 {
			return nil
		}
		return nothingIfGone(g.AddCounterByForEffect(item.Controller, ts[0].ID, game.CounterPlusOne, p.Amount))
	},
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))
