package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8a7c3718-74a1-4f26-8e06-0e17db4b0077",
		Name:         "Gilded Pinions",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gilded Pinions — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
