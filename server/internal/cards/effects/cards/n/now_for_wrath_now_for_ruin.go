package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "8141fb64-7b55-40fb-8fe4-5efd805d8fef",
		Name:         "Now for Wrath, Now for Ruin!",
		Purpose:      game.Purpose{RingTempts: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())) {
				ids = append(ids, c.InstanceID)
			}
			then := ctx.ItemTail(nowForWrathVigilanceAndTemptation)
			then.Carry = append([]uuid.UUID(nil), ids...)
			return effects.B13PutCounterOnEachThen(ctx, ids, then)
		},
	})
}

// nowForWrathVigilanceAndTemptation is "they gain vigilance until end of turn.
// The Ring tempts you", once the counters have landed; the tail's Carry is the
// creatures that were chosen. An on-disk key: never renamed, never reused.
var nowForWrathVigilanceAndTemptation = effects.CardTail("now-for-wrath/vigilance-and-temptation", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.Tail.Carry {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Now for Wrath, Now for Ruin! — vigilance"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.TheRingTemptsYou{}.Apply(ctx)
})
