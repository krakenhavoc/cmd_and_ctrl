package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vengeance Mission — Enchantment {2}{B} (Star Trek, #4973):
//
//	"Whenever a creature dies, put a quest counter on this enchantment.
//	 Target opponent loses 1 life. That player loses 2 life instead if
//	 there are seven or more quest counters on this enchantment."
//
// One trigger for any creature that dies, a token included (CR 700.4),
// with the opponent targeted as it goes on the stack. The count is read
// once the quest counter LANDS (AddCounterByKeyedForEffect's
// continuation), so the seventh counter makes that same trigger's loss
// two, and a Doubling Season pause cannot make it read the count before
// the counter is there. If the enchantment has left the battlefield by
// resolution, no counter is put anywhere and the loss reads the counters
// it last had there (CR 608.2h). A target gone by resolution loses
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	t := On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		_, ok := diedCreature(ev, g)
		return ok
	}, "Vengeance Mission — put a quest counter, then target opponent loses life", vengeanceMissionQuest)
	t.Targets = TargetPlayer("target opponent", Opponent())
	Register(Spec{
		OracleID:     "eae07535-487b-48ac-b99e-b4b62286254c",
		Name:         "Vengeance Mission",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// vengeanceMissionQuest puts the quest counter on the enchantment and
// runs the loss once it has landed; with the enchantment gone (or a new
// object), the loss alone.
func vengeanceMissionQuest(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !b15OnBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return vengeanceMissionLoss(ctx)
	}
	return g.AddCounterByKeyedForEffect(uuid.Nil, item.SourceCardID, "quest", 1, ctx.ItemTail(vengeanceMissionLossTail))
}

// vengeanceMissionLossTail is the rest of the trigger after the quest
// counter settles. An on-disk identity: never renamed, never reused.
var vengeanceMissionLossTail = CardTail("vengeance-mission/loss", func(ctx *Context, _ game.TailOutcome) error {
	return vengeanceMissionLoss(ctx)
})

// vengeanceMissionLoss is "Target opponent loses 1 life. That player
// loses 2 life instead if there are seven or more quest counters on this
// enchantment": the count as the enchantment is now, or as it last
// existed on the battlefield.
func vengeanceMissionLoss(ctx *Context) error {
	n := 1
	if info, ok := ctx.SourcePermanent(); ok && info.Counters["quest"] >= 7 {
		n = 2
	}
	return TargetPlayerLosesLife{N: n}.Apply(ctx)
}
