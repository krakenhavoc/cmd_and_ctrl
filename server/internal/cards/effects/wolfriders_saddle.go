package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d0eeb863-6c1b-4e84-9199-e77f15724626",
		Name:         "Wolfrider's Saddle",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		BlockRules:   []game.BlockRule{MaxBlockers(OnAttached(), 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wolfrider's Saddle — create a 2/2 green Wolf creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "2/2 green Wolf", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
