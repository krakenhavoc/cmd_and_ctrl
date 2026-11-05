package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guinan, Ancient Advisor — Legendary Creature — El-Aurian Advisor
// {2}{W}, 1/4 (Star Trek Commander, #4974):
//
//	"{1}, {T}: Put a +1/+1 counter on each creature you control that
//	 entered this turn. If you put two or more +1/+1 counters on creatures
//	 this way, draw a card."
//
// The creatures are the ones you control as the ability resolves that
// have an entry in this turn's tally (game.EnteredThisTurn): a creature
// that entered on an opponent's turn and is merely summoning sick does
// not count. "Two or more counters" counts what was PUT, after the CR 614
// window: a Hardened Scales makes one creature's placement two, and a
// creature that can't have counters put on it adds none. So the
// placements are chained one at a time, each settled (a CR 616 ordering
// prompt over one of them is a restore point) before the next is tried,
// with the running total carried on the tail; the draw is decided once
// the last has landed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b93915ac-744b-4b59-8f16-ae68d6d898c6",
		Name:         "Guinan, Ancient Advisor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on each creature you control that entered this turn. If you put two or more +1/+1 counters on creatures this way, draw a card.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  guinanCounters,
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
		}},
	})
}

// guinanCounters reads the set as the ability resolves and starts the
// chain.
func guinanCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == item.Controller && g.EnteredThisTurn(c.InstanceID) {
			ids = append(ids, c.InstanceID)
		}
	}
	t := ctx.ItemTail(guinanCounterNext)
	t.Carry = ids
	return guinanPlaceNext(ctx, t)
}

// guinanPlaceNext puts the next counter of the chain `t` carries, or,
// with none left, draws if the total reached two.
func guinanPlaceNext(ctx *Context, t game.KeyedTail) error {
	ids := t.Carry
	for len(ids) > 0 && !onBattlefield(ctx.Game, ids[0]) {
		ids = ids[1:]
	}
	if len(ids) == 0 {
		if t.N >= 2 {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		}
		return nil
	}
	next := t
	next.Carry = append([]uuid.UUID(nil), ids[1:]...)
	return ctx.Game.AddCounterByKeyedForEffect(ctx.Controller(), ids[0], game.CounterPlusOne, 1, next)
}

// guinanCounterNext runs behind each settled placement: it adds what
// landed to the running total and places the next. An on-disk key.
var guinanCounterNext game.TailRef

func init() {
	guinanCounterNext = CardTail("counter/guinan-ancient-advisor-next", func(ctx *Context, o game.TailOutcome) error {
		t := o.Tail
		if o.Amount > 0 {
			t.N += o.Amount
		}
		return guinanPlaceNext(ctx, t)
	})
}
