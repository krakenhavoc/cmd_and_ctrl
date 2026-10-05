package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valley Mightcaller — Creature — Frog Warrior {G}, 1/1:
//
//	"Trample
//	 Whenever another Frog, Rabbit, Raccoon, or Squirrel you control
//	 enters, put a +1/+1 counter on this creature."
//
// Subtypes are read through the effective view, so a changeling
// qualifies and the Mightcaller's own entry never counts ("another").
// The counter goes on only if the Mightcaller is still on the
// battlefield when the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "16e9c452-6288-4da8-813d-2eb6b7a538c3",
		Name:            "Valley Mightcaller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && (c.HasSubtype("Frog") || c.HasSubtype("Rabbit") || c.HasSubtype("Raccoon") || c.HasSubtype("Squirrel"))
			}, "Valley Mightcaller — put a +1/+1 counter on it", plusOneCountersOnThis(1)),
		},
	})
}
