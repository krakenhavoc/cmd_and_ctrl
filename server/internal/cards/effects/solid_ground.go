package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solid Ground — Enchantment {3}{G} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"When this enchantment enters, earthbend 3.
//	 If one or more +1/+1 counters would be put on a permanent you
//	 control, that many plus one +1/+1 counters are put on it instead."
//
// The enters trigger is the shared earthbend row (avB14WhenEntersEarthbend,
// CR 701.66a). The replacement (CR 614.1a) is Kami of Whispered Hopes'
// predicate, kamiPlusOneCounterOnYourPermanent: any permanent you control,
// including one that enters with counters (CR 122.6). Solid Ground is on
// the battlefield when its own earthbend resolves, so that land gets four
// counters. Several copies each add one, and with other counter
// replacements the affected permanent's controller picks the order
// (CR 616.1): the engine's CR 616 window does both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8609ed1a-f202-483e-9299-408ef6e84ad6",
		Name:         "Solid Ground",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{avB14WhenEntersEarthbend("Solid Ground — earthbend 3", 3)},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventCounterPlaced},
			AppliesTo: kamiPlusOneCounterOnYourPermanent,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			Label:      "Solid Ground: +1 +1/+1 counter",
		}},
	})
}
