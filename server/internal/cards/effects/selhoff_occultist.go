package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selhoff Occultist — Creature — Human Rogue {2}{U}, 2/3 (Jumpstart, #4892):
//
//	"Whenever this creature or another creature dies, target player mills
//	 a card."
//
// Blood Artist's trigger: every creature that dies, any player's and the
// Occultist itself, triggers it once, and its controller picks the
// player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d72031e1-c8cf-4e57-b15b-6a5eaacff1e2",
		Name:         "Selhoff Occultist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			},
			Targets: TargetPlayer("target player"),
			Key:     "Selhoff Occultist — target player mills a card",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  TargetPlayerMills(1),
		}},
	})
}
