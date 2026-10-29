package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Now for Wrath, Now for Ruin! — Sorcery {3}{W}:
//
//	"Put a +1/+1 counter on each creature you control. They gain
//	 vigilance until end of turn. The Ring tempts you."
//
// "They" are the creatures that got a counter, fixed as the spell
// resolves (CR 611.2c). Each counter settles before the next and
// before the rest of the sentence: a placement can stop for a
// replacement order (CR 616.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8141fb64-7b55-40fb-8fe4-5efd805d8fef",
		Name:         "Now for Wrath, Now for Ruin!",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, And(Creature(), YouControl())) {
				ids = append(ids, c.InstanceID)
			}
			then := ctx.ItemTail(nowForWrathVigilanceAndTemptation)
			then.Carry = append([]uuid.UUID(nil), ids...)
			return b13PutCounterOnEachThen(ctx, ids, then)
		},
	})
}

// nowForWrathVigilanceAndTemptation is "they gain vigilance until end of turn.
// The Ring tempts you", once the counters have landed; the tail's Carry is the
// creatures that were chosen. An on-disk key: never renamed, never reused.
var nowForWrathVigilanceAndTemptation = CardTail("now-for-wrath/vigilance-and-temptation", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.Tail.Carry {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Now for Wrath, Now for Ruin! — vigilance"}).Apply(ctx); err != nil {
			return err
		}
	}
	return TheRingTemptsYou{}.Apply(ctx)
})
