package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// MagaDrain is Maga's enters trigger: the target player loses life equal to the number
// of +1/+1 counters on the source as it resolves (CR 608.2h).
func MagaDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := CountersOnThisAsItResolves(game.CounterPlusOne)(ctx)
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
		}
	}
	return nil
}

// thenMasumaroRemove is the second half of Descendant of Masumaro's upkeep: remove a
// +1/+1 counter for each card in the target opponent's hand (the tail's N, counted
// when the ability resolved), no more than the creature has. An on-disk key: never
// renamed, never reused.
var thenMasumaroRemove = CardTail("counter/descendant-of-masumaro-remove", func(ctx *Context, o game.TailOutcome) error {
	return masumaroRemove(ctx.Game, o.Tail.Source, o.Tail.N)
})

func masumaroRemove(g *game.Game, source uuid.UUID, n int) error {
	if n <= 0 || !OnBattlefield(g, source) {
		return nil
	}
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return nil
	}
	if have := c.Counters[game.CounterPlusOne]; n > have {
		n = have
	}
	if n <= 0 {
		return nil
	}
	return g.AddCounterForEffect(source, game.CounterPlusOne, -n)
}

// DescendantOfMasumaroUpkeep is "put a +1/+1 counter on this creature for each card in
// your hand, then remove a +1/+1 counter from this creature for each card in target
// opponent's hand".
func DescendantOfMasumaroUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src := ctx.Source()
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	remove := 0
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			remove = CardsInHand(g, t.ID)
			break
		}
	}
	put := CardsInHand(g, item.Controller)
	if put <= 0 {
		return masumaroRemove(g, src, remove)
	}
	return g.AddCounterByKeyedForEffect(uuid.Nil, src, game.CounterPlusOne, put, ctx.TailN(thenMasumaroRemove, remove))
}

// CelestialKirinDestroy destroys every permanent whose mana value equals the cast
// spell's.
func CelestialKirinDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	return DestroyAllMatching{Match: ManaValueIs(obj.ManaValue)}.Apply(ctx)
}
