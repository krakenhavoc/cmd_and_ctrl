package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ragged Short Spear — Artifact — Equipment {1}{R}:
//
//	"When this Equipment enters, you may discard a card. If you do,
//	 draw two cards.
//	 Equipped creature gets +2/+0.
//	 Equip {3}"
//
// The enters trigger is Kickoff Celebrations' rummage:
// mayDiscardACardToDrawTwo, which draws only if the discard happened.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb67d0a4-864e-4f46-8538-fdd920bf0197",
		Name:         "Ragged Short Spear",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ragged Short Spear — you may discard a card. If you do, draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayDiscardACardToDrawTwo("Ragged Short Spear — discard a card to draw two?")(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
