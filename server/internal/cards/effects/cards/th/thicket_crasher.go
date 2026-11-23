package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thicket Crasher — Creature — Elemental Rhino {3}{G}, 4/3 (Core Set 2020, #5518):
//
//	"Trample (This creature can deal excess combat damage to the player or
//	 planeswalker it's attacking.)
//	 Other Elementals you control have trample."
//
// Camellia's tribal keyword grant over other Elementals you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "998e850a-b3c7-4ab2-85ac-25beef98ab20",
		Name:            "Thicket Crasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Elemental"}, Others: true, YoursOnly: true}, "trample"),
		},
	})
}
