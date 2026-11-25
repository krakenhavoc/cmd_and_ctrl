package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taunting Arbormage — Creature — Elf Wizard {2}{G}, 2/3 (Zendikar Rising,
// S82 OLD #5508):
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 When this creature enters, if it was kicked, all creatures able to
//	 block target creature this turn do so."
//
// A kicked entry triggers (CR 702.33a, 603.4) and targets a creature; the
// lure is Bloodscent's: every creature able to block it must (CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "58254569-0ba6-4ac1-9403-dc21a1b5c2e8",
		Name:          "Taunting Arbormage",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Taunting Arbormage — kicked: all creatures able to block target creature this turn do so",
					func(g *game.Game, item *game.StackItem) error {
						return effects.LureTargetForTheTurn(effects.NewContext(g, item), "Taunting Arbormage — all creatures able to block it do so")
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
