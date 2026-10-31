package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// HYDRA Disintegrator — Artifact — Equipment {4}{B} (Marvel Super Heroes
// Commander, #4455):
//
//	"When this Equipment enters, create a 2/1 black Villain creature
//	 token with menace, then attach this Equipment to it.
//	 Equipped creature gets +3/+3.
//	 Equip {4}"
//
// Dwarven Shortsword's shape: the token is created and the attach rides
// the creation's continuation (createTokenAndAttachSource). The token's
// words are the card's own, so it is read through the card's text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf16702d-9a0e-4e4c-82cb-d9b0b1e23f15",
		Name:         "HYDRA Disintegrator",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("HYDRA Disintegrator — create a 2/1 black Villain creature token with menace, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "2/1 black Villain with menace", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
