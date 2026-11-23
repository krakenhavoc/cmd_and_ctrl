package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilded Pinions — Artifact — Equipment {2} (Streets of New Capenna, #5491):
//
//	"When this Equipment enters, create a Treasure token.
//	 Equipped creature has flying.
//	 Equip {2}"
//
// The entry trigger is a plain Treasure; the flying is GrantToAttached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a7c3718-74a1-4f26-8e06-0e17db4b0077",
		Name:         "Gilded Pinions",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gilded Pinions — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
