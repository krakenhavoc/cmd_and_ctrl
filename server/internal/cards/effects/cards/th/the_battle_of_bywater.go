package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Battle of Bywater — Sorcery {1}{W}{W} (EDHREC rank 1834):
//
//	"Destroy all creatures with power 3 or greater. Then create a
//	 Food token for each creature you control. (It's an artifact with
//	 "{2}, {T}, Sacrifice this token: You gain 3 life.")"
//
// The go-wide deck's one-sided wrath: the big things die and every
// small creature you kept becomes a Food. Power is the current power
// (PowerGE reads counters and anthems), and the Food count is taken
// AFTER the destruction, so a creature of yours that survived — or
// was too small to be swept — is what pays.
//
// The batched sweep. DestroyAllMatching destroys the big creatures as
// one simultaneous event (CR 700.4), so a "whenever another creature
// dies" watcher caught in the wipe sees every death, and it leaves
// indestructible creatures on the battlefield (#446 / #470), where
// one of yours counts for a Food like any other creature you control.
//
// The Food count cannot simply read the board after the sweep. A
// commander of yours with power 3 or more is destroyed and goes to the
// graveyard like any creature (CR 903.9a, ADR 0115), whatever its
// owner then answers. The Then clause still leaves out every card in
// `swept`: those are the creatures that were destroyed this way, and
// none of them is a creature you control any more.
//
// #815: the Foods are made from the destruction's continuation, so
// they are created when that last answer arrives rather than while
// the prompt is open — and `swept` holds only what was really
// destroyed, so a creature the window saved mid-sweep is still one of
// yours and still pays a Food.
func init() {
	effects.Register(effects.Spec{
		OracleID: "a94c191d-a938-458e-bc1b-2f44fd8873a3",
		Name:     "The Battle of Bywater",
		// ADR 0126 §6: only power 3 or greater.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.PowerGE(3)), ThenTail: ctx.ItemTail(battleOfBywaterFood)}.Apply(ctx)
		},
	})
}

// battleOfBywaterFood is "then create a Food token for each creature you
// control", read after the sweep, keyed (#4253 slice 10a).
var battleOfBywaterFood = effects.MassTail("mass/the-battle-of-bywater-food", func(ctx *effects.Context, swept []game.Card, _ int) error {
	return effects.B17FoodPerCreatureYouControl(ctx, swept)
})
