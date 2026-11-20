package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Executioner's Hood — Artifact — Equipment {2} (Dark Ascension, #5571):
//
//	"Equipped creature has intimidate. (This creature can't be
//	 blocked except by artifact creatures and/or creatures that share
//	 a color with it.)
//	 Equip {2} ({2}: Attach to target creature you control. Equip
//	 only as a sorcery.)"
//
// A layer 6 grant to the equipped creature (GrantToAttached), as on Cloak
// of the Bat; equip is the shared sorcery-speed row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "041738f9-7465-4a6c-8046-a897d88cbcdd",
		Name:         "Executioner's Hood",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("intimidate")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
