package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ef4b334b-8407-4147-b434-602e59806906",
		Name:         "Hunted Phantasm",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.RavB4TargetOpponentCreates("Hunted Phantasm — target opponent creates five 1/1 red Goblins", "1/1 red Goblin", 5),
		},
	})
}
