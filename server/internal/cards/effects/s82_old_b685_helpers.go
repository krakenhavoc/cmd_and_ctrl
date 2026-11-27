package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b685_helpers.go — shared body of "put a +1/+1 counter on that creature" triggers
// (S82 OLD batch 685, #5619).

// PutPlusOneCounterOnEventCreature is the effect of a trigger that puts a +1/+1 counter on the
// creature its event names (Case of the Pilfered Proof's Detective, Ferocity's enchanted
// creature), when that object is still on the battlefield as the trigger resolves.
func PutPlusOneCounterOnEventCreature(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !OnBattlefield(g, item.Trigger.Event.CardID) {
		return nil
	}
	return AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}
