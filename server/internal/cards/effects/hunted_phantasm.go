package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunted Phantasm — Creature — Spirit {1}{U}{U}, 4/6 (Ravnica: City of
// Guilds, #5600):
//
//	"This creature can't be blocked.
//	 When this creature enters, target opponent creates five 1/1 red
//	 Goblin creature tokens."
//
// Shoreline Looter's unblockability (RestrictSelf), and Hunted Horror's
// targeted enters trigger: the opponent creates the Goblins (CR 111.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef4b334b-8407-4147-b434-602e59806906",
		Name:         "Hunted Phantasm",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			ravB4TargetOpponentCreates("Hunted Phantasm — target opponent creates five 1/1 red Goblins", "1/1 red Goblin", 5),
		},
	})
}
