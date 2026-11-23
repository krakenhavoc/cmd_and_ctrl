package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Haruspex — Creature — Human Wizard {2}{B}, 3/2:
//
//	"Morph {B}
//	 Whenever another nontoken creature you control dies, draw a card."
//
// Morph is the face-down alternative cost (CR 702.37). The dies trigger is
// Body Launderer's test (another nontoken creature, read by last-known
// information, under this controller) with a draw; the card has no "you may",
// so the draw is unconditional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "72aef1d8-11df-4bc2-af83-907b05df73a8",
		Name:             "Grim Haruspex",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{B}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B19AnotherNontokenCreatureYouControlDied(ev, source, g)
			}, "Grim Haruspex — draw a card", effects.Do(effects.DrawCards{N: 1})),
				game.Purpose{Draws: 1, DeathPayoff: true}),
		},
	})
}
