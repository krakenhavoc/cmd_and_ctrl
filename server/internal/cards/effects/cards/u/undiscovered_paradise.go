package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undiscovered Paradise — Land (Visions, S82 older sets batch 792, #5631):
//
//	"{T}: Add one mana of any color. During your next untap step, as you untap
//	 your permanents, return this land to its owner's hand."
//
// The mana ability's rider writes the engine's untap-step record (#6218): it is not
// a delayed trigger, because the untap step has no priority (CR 502.4), but an
// action taken as the controller's permanents untap (CR 502.3). The record names
// the land as the object it is now (CR 400.7), so a land that has left or been
// replaced is not returned. A land carrying a hand-written rider is never
// auto-tapped, which suits a land that leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76c33d54-ce55-400e-bec5-79d33a5a20fb",
		Name:         "Undiscovered Paradise",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(1),
			Label:    "Add one mana of any color. During your next untap step, as you untap your permanents, return this land to its owner's hand.",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				g.ReturnAtNextUntapForEffect(source, controller, "Undiscovered Paradise — return this land to its owner's hand")
				return nil
			},
		}},
	})
}
