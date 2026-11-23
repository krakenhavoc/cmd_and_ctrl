package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolfrider's Saddle — Artifact — Equipment {3}{G} (Foundations Jumpstart):
//
//	"When this Equipment enters, create a 2/2 green Wolf creature token,
//	 then attach this Equipment to it.
//	 Equipped creature gets +1/+1 and can't be blocked by more than one
//	 creature.
//	 Equip {3}"
//
// Dwarven Shortsword's enters trigger with a Wolf (createTokenAndAttachSource,
// described), and Rope's block limit on the equipped creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0eeb863-6c1b-4e84-9199-e77f15724626",
		Name:         "Wolfrider's Saddle",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnAttached(), 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wolfrider's Saddle — create a 2/2 green Wolf creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "2/2 green Wolf", true)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
