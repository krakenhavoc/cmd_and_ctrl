package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cf16702d-9a0e-4e4c-82cb-d9b0b1e23f15",
		Name:         "HYDRA Disintegrator",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("HYDRA Disintegrator — create a 2/1 black Villain creature token with menace, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "2/1 black Villain with menace", true)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
