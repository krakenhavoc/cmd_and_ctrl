package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crusading Knight — Creature — Human Knight {2}{W}{W}, 2/2 (Invasion, S82 OLD
// #5616):
//
//	"Protection from black
//	 This creature gets +1/+1 for each Swamp your opponents control."
//
// Protection from black is printed (CR 702.16). The bonus is a layer 7c static
// that counts the Swamps every opponent of its controller controls, read through
// the layers as they stand (CR 604.2, 613.4c): it grows and shrinks as lands
// come and go, and your own Swamps never count. Swamp is a word of the
// Knight's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "59385809-3223-4bc2-a824-9bab257c917a",
		Name:            "Crusading Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Static: []game.StaticAbility{
			effects.OldB652PumpPerOpponentsPermanent("Swamp", 1, 1, "This creature gets +1/+1 for each Swamp your opponents control."),
		},
	})
}
