package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Draugr Thought-Thief — Creature — Zombie Rogue {2}{U}, 3/2:
//
//	"When this creature enters, look at the top card of target player's
//	 library. You may put that card into their graveyard."
func init() {
	enters := effects.WhenThisEnters("Draugr Thought-Thief — look at the top card of target player's library; you may put it into their graveyard",
		effects.S82BnsB28DraugrLook)
	enters.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:     "0952cfaf-8031-4df7-b276-dcb0b8f19ab7",
		Name:         "Draugr Thought-Thief",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
