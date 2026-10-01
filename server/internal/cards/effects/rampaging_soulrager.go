package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampaging Soulrager — Creature — Spirit {2}{R}, 1/4:
//
//	"This creature gets +3/+0 as long as there are two or more unlocked
//	 doors among Rooms you control."
//
// A layer 7c static on itself whose condition is UnlockedDoorsYouControl
// (rooms.go). The layer cache drops on every door lock and unlock
// (EventDoorUnlocked / EventDoorLocked), and on a Room entering or
// leaving, so the bonus switches on and off with the board. It counts
// doors, not Rooms: one fully unlocked Room is enough.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6d8e6cb-0878-4f2c-9a89-ce84cb4f5bbc",
		Name:         "Rampaging Soulrager",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && g != nil &&
					UnlockedDoorsYouControl(g, source.Controller) >= 2
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 3
			},
		}},
	})
}
