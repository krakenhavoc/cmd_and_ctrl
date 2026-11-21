package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ondu Cleric — Creature — Kor Cleric Ally {1}{W}, 1/1 (Zendikar, #5583):
//
//	"Whenever this creature or another Ally you control enters, you may
//	 gain life equal to the number of Allies you control."
//
// The Zendikar Ally trigger (zen_b2_helpers.go). The Allies are counted
// as the ability resolves, every permanent with the subtype and the
// Cleric itself included while it is still there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4232466-dd6a-49bf-be6c-95905c3ded17",
		Name:         "Ondu Cleric",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ZenB2AllyTrigger("Ondu Cleric — you may gain life equal to the number of Allies you control",
				"Gain life equal to the number of Allies you control?",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GainLife{Amount: effects.ZenB2AlliesYouControl(ctx)}.Apply(ctx)
				}),
		},
	})
}
