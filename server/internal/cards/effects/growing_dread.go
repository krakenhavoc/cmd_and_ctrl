package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Growing Dread — Enchantment {G}{U}:
//
//	"Flash
//	 When this enchantment enters, manifest dread.
//	 Whenever you turn a permanent face up, put a +1/+1 counter on it."
//
// The counter goes on the permanent as it is now, face up: EventTurnedFaceUp
// fires after the turn, and CR 708.8 says it is the same object. If it
// left the battlefield before the trigger resolved there is nothing to
// put a counter on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aa1d8bfa-729a-4e10-8069-57a1d4f4392b",
		Name:            "Growing Dread",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Growing Dread — manifest dread", Do(ManifestDread{})),
			On(game.EventTurnedFaceUp, ByYou, "Growing Dread — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !b15OnBattlefield(g, item.Trigger.Event.CardID) {
						return nil
					}
					return AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
