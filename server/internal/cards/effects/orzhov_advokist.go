package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orzhov Advokist — Creature — Human Advisor {2}{W}, 1/4:
//
//	"At the beginning of your upkeep, each player may put two +1/+1
//	 counters on a creature they control. If a player does, creatures that
//	 player controls can't attack you or planeswalkers you control until
//	 your next turn."
//
// The fan-out is cmdr2S34EachPlayerMayPutCounters (Agitator Ant's). For each
// player whose creature really got the counters, a live-rule record
// (game.ScopeYourCreatures, the "you" being that player) bars every creature
// the player controls from attacking the Advokist's controller or their
// planeswalkers until the Advokist controller's next turn. It changes no
// characteristic, so it reaches a creature the player controls later too
// (CR 611.2c). The Advokist's controller is a player too: their own creature
// is barred from attacking themselves, which CR 506.2 already does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5e9942e-096f-433f-8668-4b594f3c1b02",
		Name:         "Orzhov Advokist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Orzhov Advokist — each player may put two +1/+1 counters on a creature they control; those players' creatures can't attack you",
				func(g *game.Game, item *game.StackItem) error {
					return cmdr2S34EachPlayerMayPutCounters(NewContext(g, item), cmdr2S34RiderCantAtk)
				}),
		},
	})
}
