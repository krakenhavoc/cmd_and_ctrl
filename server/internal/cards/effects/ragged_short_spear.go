package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bb67d0a4-864e-4f46-8538-fdd920bf0197",
		Name:         "Ragged Short Spear",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ragged Short Spear — you may discard a card. If you do, draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return mayDiscardACardToDrawTwo("Ragged Short Spear — discard a card to draw two?")(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
