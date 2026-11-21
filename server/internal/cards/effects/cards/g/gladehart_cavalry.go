package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gladehart Cavalry — Creature — Elf Knight {5}{G}{G}, 6/6 (Oath of the Gatewatch, #5545):
//
//	"When this creature enters, support 6. (Put a +1/+1 counter on each
//	 of up to six other target creatures.)
//	 Whenever a creature you control with a +1/+1 counter on it dies,
//	 you gain 2 life."
//
// Support from a permanent targets OTHER creatures (CR 701.41a), as
// Skyboon Evangelist's support 6. The life gain is Laid to Rest's second
// trigger: the dead creature's counters are read as it last existed (CR
// 603.10a), and "you control" is the controller it died under.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d937ea3-bfd4-4cdf-b722-08f2b99ea2a5",
		Name:         "Gladehart Cavalry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersSupport("Gladehart Cavalry", 6, true),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Gladehart Cavalry — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
