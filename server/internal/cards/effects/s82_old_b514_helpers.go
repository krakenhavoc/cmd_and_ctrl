package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b514_helpers.go — shared bodies for S82 OLD batch 514 (Shards of
// Alara, #5588).

// OldB514AnotherCreatureDiedMayCounter is "Whenever another creature dies, you
// may put a +1/+1 counter on this creature" (Scavenger Drake, Algae Gharial):
// the trigger, optional, with the counter put on the source if it is still on
// the battlefield.
func OldB514AnotherCreatureDiedMayCounter(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventLTB},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return B15AnotherCreatureDied(ev, source, g)
		},
		OptionalPrompt: &game.TriggerOptionalPrompt{Question: name + " — put a +1/+1 counter on it?"},
		Key:            name + " — you may put a +1/+1 counter on it",
		Effect:         PutCounterOnSelf,
	}
}

// OldB514DevouredTrigger is the "When this creature enters, [effect] for each
// creature it devoured" half of a devour creature (Tar Fiend): the ETB trigger
// reads Card.Devoured as it is built (CR 702.82b) and carries it as the item's
// Amount for the effect to use.
func OldB514DevouredTrigger(label string, effect Effect) game.TriggeredAbility {
	t := WhenThisEnters(label, effect)
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if source != nil {
			item.Params.Amount = source.Devoured
		}
		return item
	}
	return t
}

// OldB514TargetPlayerDiscardsItemAmount is TargetPlayerDiscards for a count
// the trigger carried as its Amount: the target player picks the cards (CR
// 701.9a). A target that left the game, or an Amount of zero, discards nothing.
func OldB514TargetPlayerDiscardsItemAmount(g *game.Game, item *game.StackItem) error {
	return TargetPlayerDiscards(item.Params.Amount)(g, item)
}
