package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bounding Felidar — Creature — Cat Beast Mount {5}{W}:
//
//	"Whenever this creature attacks while saddled, put a +1/+1 counter
//	 on each other creature you control. You gain 1 life for each of
//	 those creatures.
//	 Saddle 2"
//
// "Those creatures" are the ones the counter went on, fixed before the
// first counter lands; the life is gained once every counter has settled
// (a Doubling Season board pauses a placement on a CR 616 prompt), as
// Finneas, Ace Archer's draw is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b00f316d-7c82-40e6-80e5-9f36434590a3",
		Name:         "Bounding Felidar",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(2)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Bounding Felidar — a +1/+1 counter on each other creature you control, gain 1 life for each",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != item.SourceCardID {
							ids = append(ids, c.InstanceID)
						}
					}
					ctx := NewContext(g, item)
					return b13PutCounterOnEachThen(ctx, ids, ctx.TailN(boundingFelidarGainLife, len(ids)))
				}),
		},
	})
}

// boundingFelidarGainLife is "gain 1 life for each" creature that got a
// counter, once they have landed; the tail's N is how many. An on-disk key:
// never renamed, never reused.
var boundingFelidarGainLife = CardTail("bounding-felidar/gain-life", func(ctx *Context, o game.TailOutcome) error {
	return GainLife{Player: ctx.Controller(), Amount: o.Tail.N}.Apply(ctx)
})
