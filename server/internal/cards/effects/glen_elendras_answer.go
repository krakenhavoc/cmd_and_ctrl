package effects

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glen Elendra's Answer — {2}{U}{U} Instant (Lorwyn Eclipsed, #4458):
//
//	"This spell can't be countered.
//	 Counter all spells your opponents control and all abilities your
//	 opponents control. Create a 1/1 blue and black Faerie creature token
//	 with flying for each spell and ability countered this way."
//
// The uncounterable half is Spec.CantBeCountered (Counterflux). The sweep
// is every spell, activated ability and triggered ability on the stack an
// opponent controls as the Answer resolves (CR 608.2h: the set is read
// once, so nothing that arrives during the sweep joins it). Mana abilities
// never use the stack (CR 605.3b), so they are never there to counter
// (the ruling), and static abilities are not on the stack either.
//
// "Countered this way" is counted, not assumed: each object is countered
// in turn through game.CounterKeyedForEffect, whose continuation is told
// whether it really was. A spell or ability that can't be countered is
// still on the stack and makes no Faerie; a counter that a replacement
// turned into an exile (Guile) counters nothing and makes none; a
// countered spell that went somewhere other than its graveyard (a
// flashback spell, a commander) was countered and makes one. The sweep is
// data (a keyed chain carrying the objects and the count), so a counter
// paused on a CR 616.1 ordering prompt is a restore point, and the tokens
// are made once the last counter has settled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "11ed8209-2184-4416-a8f0-72de00599574",
		Name:            "Glen Elendra's Answer",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var objects []*game.StackItem
			for id, it := range ctx.Game.StackMeta {
				if it == nil || id == item.ID || it.Controller == ctx.Controller() {
					continue
				}
				switch it.Kind {
				case game.StackItemSpell, game.StackItemActivated, game.StackItemTriggered:
					objects = append(objects, it)
				}
			}
			// Top of the stack first, for a deterministic order.
			sort.Slice(objects, func(i, j int) bool { return objects[i].Seq > objects[j].Seq })
			chain := ctx.ItemTail(glenElendrasAnswerNext)
			for _, it := range objects {
				chain.Carry = append(chain.Carry, it.ID)
			}
			chain.Nums = []int{0, 0}
			return chain.Run(ctx.Game, nil)
		},
	})
}

// glenElendrasAnswerNext counters the next object of the sweep and adds up
// what was countered. Carry is the swept objects in order; Nums are the
// next one's index and the count so far. The previous counter's outcome
// is IDs (the object when it was countered). When the sweep is done it
// makes one Faerie for each object countered. An on-disk key.
var glenElendrasAnswerNext = CardTail("then/glen-elendras-answer-counter-next", func(ctx *Context, o game.TailOutcome) error {
	t := o.Tail
	if len(t.Nums) < 2 {
		return nil
	}
	next, countered := t.Nums[0], t.Nums[1]+len(o.IDs)
	if next < len(t.Carry) {
		step := t
		step.Nums = []int{next + 1, countered}
		return ctx.Game.CounterKeyedForEffect(t.Carry[next], step)
	}
	if countered == 0 {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 blue and black Faerie with flying"), N: countered}.Apply(ctx)
})
