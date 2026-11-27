package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wu Scout — Creature — Human Soldier Scout {1}{U}, 1/1
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"When this creature enters, look at target opponent's hand."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef577d7d-7a24-4bd3-b86b-f69e37bb2ce7",
		Name:            "Wu Scout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Wu Scout — look at target opponent's hand",
				func(g *game.Game, item *game.StackItem) error {
					effects.LookAtTargetPlayersHand(item, effects.NewContext(g, item))
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
