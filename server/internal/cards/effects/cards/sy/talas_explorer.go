package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Talas Explorer — Creature — Human Pirate Scout {1}{U}, 1/1 (Portal Second Age, S82 older sets batch 734, #5625):
//
//	"Flying
//	 When this creature enters, look at target opponent's hand."
//
// The opponent is the target; the look is Wu Scout's; nothing is revealed to anyone else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da359465-f823-4989-9acd-249a2a1f3f17",
		Name:            "Talas Explorer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Talas Explorer — look at target opponent's hand",
				func(g *game.Game, item *game.StackItem) error {
					effects.LookAtTargetPlayersHand(item, effects.NewContext(g, item))
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
