package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitesail — Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+0 and has flying.
//	 Equip {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b079f9db-974d-4525-a894-57b754ba9dcc",
		Name:         "Kitesail",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
