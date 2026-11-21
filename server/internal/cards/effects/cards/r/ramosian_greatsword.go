package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ramosian Greatsword — Artifact — Equipment {4}{R} (March of the
// Machine): "Convoke. Equipped creature gets +3/+1 and has trample.
// Equip {2}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c53a177-c8c8-4b9f-a221-f205da67daa5",
		Name:         "Ramosian Greatsword",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 1),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
