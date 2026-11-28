package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wu Spy — Creature — Human Soldier Rogue {1}{U}, 1/1 (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"When this creature enters, look at the top two cards of target player's library. Put one of them into their graveyard."
//
// Thoughtpicker Witch's look with a graveyard in place of exile
// (S82OldB714WuSpyLook): only the controller looks, the pick is
// mandatory, the other card stays on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b681770-eea3-459c-ad5b-138834213f8c",
		Name:         "Wu Spy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Wu Spy — look at the top two cards of target player's library, put one into their graveyard", effects.S82OldB714WuSpyLook),
				effects.TargetPlayer("target player")),
		},
	})
}
