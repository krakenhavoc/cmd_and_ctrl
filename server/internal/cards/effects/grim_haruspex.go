package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "72aef1d8-11df-4bc2-af83-907b05df73a8",
		Name:             "Grim Haruspex",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{B}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b19AnotherNontokenCreatureYouControlDied(ev, source, g)
			}, "Grim Haruspex — draw a card", Do(DrawCards{N: 1})),
				game.Purpose{Draws: 1, DeathPayoff: true}),
		},
	})
}
