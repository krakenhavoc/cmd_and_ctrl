package cardsst

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stumpsquall Hydra — Creature — Hydra {X}{G}{G}{G}, 1/1 (Commander
// Legends, #5507):
//
//	"When this creature enters, distribute X +1/+1 counters among it and
//	 any number of commanders."
//
// X is the X announced for the spell, bound to the trigger as it is put on
// the stack (CR 107.3m). The ability has no target (CR 115.1), so the
// distribution is made as it resolves: each commander on the battlefield,
// whoever controls it, is asked in turn how many of the remaining counters
// it gets (zero leaves it out), and the Hydra, which is always among the
// recipients, takes what is left, at least one. A commander that has left
// the battlefield by then is skipped and a Hydra that has left takes
// nothing, so the last commander asked takes what remains (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b157b9a-92da-4054-8441-04fb4e415d77",
		Name:         "Stumpsquall Hydra",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       "Stumpsquall Hydra — distribute X +1/+1 counters among it and any number of commanders",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Stumpsquall Hydra — distribute X +1/+1 counters among it and any number of commanders")
				item.XValue = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var queue []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCommander && c.InstanceID != ctx.Source() {
						queue = append(queue, c.InstanceID)
					}
				}
				if _, ok := g.PermanentRefForEffect(ctx.Source()); ok {
					queue = append(queue, ctx.Source())
				}
				return stumpsquallHydraNext(ctx, ctx.X(), queue)
			},
		}},
	})
}

// stumpsquallHydraNext hands out the counters still to place. queue is the
// recipients not yet asked; the Hydra, when it is there, is last.
func stumpsquallHydraNext(ctx *effects.Context, remaining int, queue []uuid.UUID) error {
	g := ctx.Game
	for remaining > 0 && len(queue) > 0 {
		id := queue[0]
		if _, ok := g.PermanentRefForEffect(id); !ok {
			queue = queue[1:]
			continue
		}
		last := len(queue) == 1
		if last {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: remaining}).Apply(ctx); err != nil {
				return err
			}
			return nil
		}
		// The Hydra, last in the queue, must keep one counter.
		max := remaining
		if queue[len(queue)-1] == ctx.Source() {
			max--
		}
		if max <= 0 {
			queue = queue[1:]
			continue
		}
		name := "a commander"
		if c, ok := g.LookupCardForEffect(id); ok {
			name = c.Name
		}
		return effects.ChooseNumber{
			Min:      0,
			Max:      max,
			Question: "Stumpsquall Hydra — how many of the " + strconv.Itoa(remaining) + " +1/+1 counters go on " + name + "?",
			Unit:     game.PayAmountCounters,
			Then:     stumpsquallHydraGive,
			Carry:    queue,
			Nums:     []int{remaining},
		}.Apply(ctx)
	}
	return nil
}

// stumpsquallHydraGive puts the chosen number on the recipient asked and
// carries on with the rest. An on-disk key: never renamed, never reused.
var stumpsquallHydraGive game.NumberThen

// The recipient chain refers back to itself, so the key is registered in init.
func init() {
	stumpsquallHydraGive = effects.NumberThen("number/old-b081-stumpsquall-hydra-give", func(ctx *effects.Context, r game.NumberAnswer) error {
		if len(r.Carry) == 0 || len(r.Nums) == 0 {
			return nil
		}
		id, rest := r.Carry[0], r.Carry[1:]
		if r.Number > 0 {
			if _, ok := ctx.Game.PermanentRefForEffect(id); ok {
				if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: r.Number}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		return stumpsquallHydraNext(ctx, r.Nums[0]-r.Number, rest)
	})
}
