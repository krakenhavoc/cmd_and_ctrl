package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Draugr Thought-Thief — Creature — Zombie Rogue {2}{U}, 3/2:
//
//	"When this creature enters, look at the top card of target player's
//	 library. You may put that card into their graveyard."
func init() {
	enters := WhenThisEnters("Draugr Thought-Thief — look at the top card of target player's library; you may put it into their graveyard",
		s82BnsB28DraugrLook)
	enters.Targets = TargetPlayer("target player")
	Register(Spec{
		OracleID:     "0952cfaf-8031-4df7-b276-dcb0b8f19ab7",
		Name:         "Draugr Thought-Thief",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
