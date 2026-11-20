package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Experimental Armor — Artifact — Equipment {2}{R} (Marvel Super Heroes
// Commander, #4455): "Equipped creature gets +1/+1 and has flying and
// haste. Equip {2}". The pump and keyword grants are the ordinary
// attachment statics, as Batterbone's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7ed3cd6-d138-4c12-93bb-91629492a5fa",
		Name:         "Experimental Armor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1), effects.GrantToAttached("flying", "haste")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
