package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Study the Classics — Sorcery {2}{G} (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"Put a +1/+1 counter on target creature, then double the number of
//	 +1/+1 counters on it. You gain life equal to the number of +1/+1
//	 counters on that creature."
//
// Invigorating Surge's two placements on any creature, in printed order:
// one counter, then as many more as the creature now has (the ruling: to
// double is to put that many on, so a Hardened Scales sees both), then
// the life, read from the counters on it after the doubling has landed.
// Each step is a registered tail behind the one before it, so a counter
// placement that stops on a CR 616 ordering prompt is a restore point and
// the next step reads the count only once the placement is done (#1290).
// A target that left in response fizzles the spell (CR 608.2b): no
// counters and no life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bfff503e-3d16-436c-afe5-ec6a03794725",
		Name:         "Study the Classics",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			then := ctx.ItemTail(studyTheClassicsDouble)
			then.Carry = []uuid.UUID{id}
			return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, id, game.CounterPlusOne, 1, then)
		},
	})
}

// studyTheClassicsDouble doubles the +1/+1 counters on Carry[0] once the
// first one has landed, then hands on to the life gain. An on-disk key:
// never renamed, never reused.
var studyTheClassicsDouble = CardTail("study-the-classics/double", func(ctx *Context, o game.TailOutcome) error {
	id := carryFirst(o.Tail)
	n := 0
	if onBattlefield(ctx.Game, id) {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			n = c.Counters[game.CounterPlusOne]
		}
	}
	if n <= 0 {
		return studyTheClassicsGainFor(ctx, id)
	}
	then := ctx.ItemTail(studyTheClassicsGain)
	then.Carry = []uuid.UUID{id}
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, id, game.CounterPlusOne, n, then)
})

// studyTheClassicsGain is "you gain life equal to the number of +1/+1
// counters on that creature", once the doubling has landed. An on-disk
// key: never renamed, never reused.
var studyTheClassicsGain = CardTail("study-the-classics/gain", func(ctx *Context, o game.TailOutcome) error {
	return studyTheClassicsGainFor(ctx, carryFirst(o.Tail))
})

func studyTheClassicsGainFor(ctx *Context, id uuid.UUID) error {
	if !onBattlefield(ctx.Game, id) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: c.Counters[game.CounterPlusOne]}.Apply(ctx)
}
