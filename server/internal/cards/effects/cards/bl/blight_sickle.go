package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blight Sickle — Artifact — Equipment {2} (Shadowmoor, #5591):
//
//	"Equipped creature gets +1/+0 and has wither. (It deals damage to
//	 creatures in the form of -1/-1 counters.)
//	 Equip {2}"
//
// Boots of Speed's shape: PumpAttached for the +1/+0 and GrantToAttached for
// the keyword, both layered statics that end when the Sickle moves or
// leaves. Equip is the shared sorcery-speed attach ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b95a5ed-a6b2-4044-9273-2021d27794e9",
		Name:         "Blight Sickle",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0), effects.GrantToAttached("wither")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
