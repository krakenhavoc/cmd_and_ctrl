package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honored Dreyleader — Creature — Squirrel Warrior {2}{G}, 1/1
// (EDHREC rank 3172):
//
//	"Trample
//	 When this creature enters, put a +1/+1 counter on it for each
//	 other Squirrel and/or Food you control.
//	 Whenever another Squirrel or Food you control enters, put a
//	 +1/+1 counter on this creature."
//
// The Squirrel deck's scaling body. Trample rides PrintedKeywords.
// The entry trigger counts the other Squirrels and Foods as it
// resolves — one counter per permanent, so a permanent that is both
// counts once, as the "and/or" says. The growth trigger fires once
// per Squirrel or Food entering under the controller's control, a
// token's entry included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "44bffa3e-4d31-4f9b-829c-e2066ab2f641",
		Name:            "Honored Dreyleader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Honored Dreyleader — a +1/+1 counter per other Squirrel or Food you control", effects.B30CountersForOtherSquirrelsAndFood),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B30AnotherSquirrelOrFoodYouControlEntered(ev, source, g)
			}, "Honored Dreyleader — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
