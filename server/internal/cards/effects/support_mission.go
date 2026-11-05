package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Support Mission — {2}{G} Enchantment (Star Trek, #4973):
//
//	"Whenever a creature you control enters, put a quest counter on this
//	 enchantment. Put a +1/+1 counter on that creature. If this
//	 enchantment has five or more quest counters on it, instead put two
//	 +1/+1 counters on that creature."
//
// One ability, read in order as it resolves (CR 608.2c): the quest counter
// goes on first, so the fifth creature to enter is the first to get two.
// The count is read after that counter, from the enchantment as it last
// existed if it has left the battlefield (CR 608.2h). "That creature" is
// the object that entered; one that has left is a new object and gets
// nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59a4c4cf-07b4-4009-bbea-7da7b31f09a1",
		Name:         "Support Mission",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl,
				"Support Mission — put a quest counter on this enchantment and +1/+1 counters on that creature",
				supportMissionResolve),
		},
	})
}

func supportMissionResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}).Apply(ctx); err != nil {
		return err
	}
	if item.Trigger == nil {
		return nil
	}
	entered := item.Trigger.Event.CardID
	if !onBattlefield(g, entered) {
		return nil
	}
	n := 1
	if trekB5CountersOnThis(g, item.SourceCardID, "quest") >= 5 {
		n = 2
	}
	return AddCounter{Target: entered, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}
