package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hero of Leina Tower — {G} Creature — Human Warrior 1/1 (Born of the Gods):
//
//	"Heroic — Whenever you cast a spell that targets this creature, you may pay {X}. If you
//	 do, put X +1/+1 counters on this creature."
//
// X is chosen as the trigger resolves (CR 608.2d), as Mortarion's is; paying nothing
// declines, and the counters go on the Hero only while it is still the object that
// triggered (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae15a96d-ce88-4cb9-a353-439387ca3d02",
		Name:         "Hero of Leina Tower",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Hero of Leina Tower — you may pay {X}; if you do, put X +1/+1 counters on it (heroic)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayX{
						Cost:  "{X}",
						Label: "Hero of Leina Tower",
						Buys:  "put X +1/+1 counters on this creature",
						Goal:  effects.XAsHighAsYouCan,
						Then:  heroOfLeinaTowerPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// heroOfLeinaTowerPaid is the paid branch, keyed: an on-disk identity, never renamed, never
// reused.
var heroOfLeinaTowerPaid = effects.PayUnlessBranches("pay-unless/hero-of-leina-tower-counters",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 || ctx.Item == nil || effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Item.SourceCardID) {
			return nil
		}
		return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: r.Number}.Apply(ctx)
	}, nil)
