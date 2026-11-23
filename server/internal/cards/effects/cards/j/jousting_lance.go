package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jousting Lance — Artifact — Equipment {2} (Jumpstart 2022, #5948):
//
//	"Equipped creature gets +2/+0.
//	 During your turn, equipped creature has first strike.
//	 Equip {3}"
//
// "Your turn" is the Equipment controller's turn, read live on each layer
// pass, so first strike comes and goes with the turn (CR 611.3a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab747483-2aa1-4b25-b2fa-cf7372a1b1dc",
		Name:         "Jousting Lance",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttachedWhile(effects.DuringSourceControllersTurn, "first strike"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
