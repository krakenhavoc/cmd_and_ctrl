package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunfall — Sorcery {3}{W}{W} (Reality Fracture, tracker #2795):
//
//	"Exile all creatures. Incubate X, where X is the number of creatures
//	 exiled this way."
//
// ExileAllMatching's Then runs from a continuation once every creature
// has finished moving (a commander's CR 903.9 prompt holds it), with the
// count of cards that really reached exile; the keyword action (CR
// 701.53a, ADR 0185) is then given that count. The Incubator is made
// after the sweep, so it is not itself exiled, and it enters with X
// +1/+1 counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb3f5097-0da0-458c-8508-60823567e2da",
		Name:         "Sunfall",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileAllMatching{Match: effects.Creature(), ThenTail: ctx.ItemTail(sunfallIncubate)}.Apply(ctx)
		},
	})
}

// sunfallIncubate is "incubate X, where X is the number of creatures
// exiled this way", keyed (#4253 slice 10a).
var sunfallIncubate = effects.MassTail("mass/sunfall-incubate", func(ctx *effects.Context, _ []game.Card, exiled int) error {
	return effects.Incubate{N: exiled}.Apply(ctx)
})
