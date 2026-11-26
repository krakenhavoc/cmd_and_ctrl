package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB541MinusCounterOnEach puts one -1/-1 counter on every listed
// permanent still on the battlefield, in the given order. Each placement is
// chained from the previous one's own tail, so a CR 616 ordering prompt over
// any one of them pauses the rest (Lockjaw Snapper).
func OldB541MinusCounterOnEach(ctx *Context, ids []uuid.UUID) error {
	for len(ids) > 0 {
		if z := ctx.Game.FindCardZoneForEffect(ids[0]); z != nil && z.Kind == game.ZoneBattlefield {
			break
		}
		ids = ids[1:]
	}
	if len(ids) == 0 {
		return nil
	}
	next := ctx.ItemTail(oldB541MinusCounterNext)
	next.Carry = append([]uuid.UUID(nil), ids[1:]...)
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, ids[0], game.CounterMinusOne, 1, next)
}

// oldB541MinusCounterNext is the rest of an OldB541MinusCounterOnEach list
// behind a settled placement. An on-disk key: never renamed, never reused.
var oldB541MinusCounterNext game.TailRef

var _ = registerOldB541MinusCounterNext()

func registerOldB541MinusCounterNext() bool {
	oldB541MinusCounterNext = CardTail("old-b541/minus-counter-on-each-next", func(ctx *Context, o game.TailOutcome) error {
		return OldB541MinusCounterOnEach(ctx, o.Tail.Carry)
	})
	return true
}
