package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kate Stewart — Legendary Creature — Human Scientist {1}{U}{R}{W}, 3/3
// (Doctor Who, #5948):
//
//	"Whenever you put one or more time counters on a permanent you control,
//	 create a 1/1 white Soldier creature token.
//	 Whenever Kate Stewart attacks, you may pay {8}. If you do, attacking
//	 creatures get +X/+X until end of turn, where X is the number of time
//	 counters among permanents you control."
//
// The first trigger is once per permanent per batch of time counters YOU put
// on a permanent you control (CR 603.2c; a card in exile with suspend
// counters is not a permanent, so a suspend does not trigger it). The
// second asks for {8} as Kate attacks and, if it is paid, gives every
// attacking creature (any controller's) +X/+X until end of turn, X read as
// the ability resolves, from the time counters on the permanents you control
// then.
//
// No simplification.
func init() {
	soldier := Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1})
	first := OncePerBatch(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if source == nil || ev.Label != game.CounterTime {
			return false
		}
		if _, ok := countersPutBy(ev, source.Controller, g, true); !ok {
			return false
		}
		if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.Target)
		return ok && c.Controller == source.Controller
	}, "Kate Stewart — create a 1/1 white Soldier creature token", soldier))
	first.BatchKey = counterTargetBatchKey
	Register(Spec{
		OracleID:     "c1b9475e-a272-4820-b5f0-25f7a94ff834",
		Name:         "Kate Stewart",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			first,
			WheneverThisAttacks("Kate Stewart — you may pay {8}: attacking creatures get +X/+X until end of turn, where X is the number of time counters among permanents you control",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{8}",
						Question: "Kate Stewart — pay {8} to give attacking creatures +X/+X until end of turn, where X is the number of time counters among permanents you control?",
						Then:     kateStewartPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// kateStewartTimeCounters is the number of time counters among permanents
// controller controls.
func kateStewartTimeCounters(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller {
			n += c.Counters[game.CounterTime]
		}
	}
	return n
}

var kateStewartPaid = PayUnlessBranches("pay-unless/kate-stewart-pump",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		x := kateStewartTimeCounters(ctx.Game, ctx.Controller())
		return BoostUntilEOT{
			Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && c.AttackingTarget != uuid.Nil
			},
			Power: x, Toughness: x,
			Label: "Kate Stewart — attacking creatures get +X/+X until end of turn",
		}.Apply(ctx)
	}, nil)
