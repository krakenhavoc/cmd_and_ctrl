package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Buzzard-Wasp Colony — Creature — Bird Insect {3}{B}, 2/2:
//
//	"Flying
//	 When this creature enters, you may sacrifice an artifact or creature.
//	 If you do, draw a card.
//	 Whenever another creature you control dies, if it had counters on it,
//	 put its counters on this creature."
//
// The entry is Khan, Vengeful Heart's may-sacrifice, asked at resolution,
// without "another": the Colony itself may be the sacrifice. The draw is
// Automated Warfare System's "if you do".
//
// The death trigger is Erebos, Bleak-Hearted's "another creature you
// control dies" with The Ozolith's body: it puts the same number of each
// kind of counter the creature had as it last existed onto this creature
// (CR 122.8, the rulings), -1/-1 counters included, so the Colony may die
// of them. Its "if it had counters" reads last-known information, which
// cannot change before resolution (CR 603.4). A Colony that left and came
// back is a new object and gets nothing (CR 400.7).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "29fd050c-dd1d-4e1b-b3cc-f7d8d18e4dcb",
		Name:            "Buzzard-Wasp Colony",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Buzzard-Wasp Colony — you may sacrifice an artifact or creature to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Buzzard-Wasp Colony — sacrifice an artifact or creature to draw a card?",
						effects.Or(effects.Artifact(), effects.Creature()), false, effects.AutomatedWarfareSystemDraw)
				}),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B39AnotherCreatureYouControlDied(ev, source, g) && len(g.LastKnownCountersForEffect(ev.CardID)) > 0
			}, "Buzzard-Wasp Colony — put its counters on this creature", effects.OzolithCollectCounters),
		},
	})
}
