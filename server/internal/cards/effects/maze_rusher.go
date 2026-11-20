package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maze Rusher — Creature — Elemental {5}{R}, 6/3:
//
//	"Haste
//	 Multicolored creatures you control have haste."
//
// The printed haste plus a layer-6 grant (CR 613.1f) of haste to every
// creature its controller controls that is two or more colours
// (CR 105.2b), read through the layers like Rienne's multicolored anthem.
// A monocolored or colourless creature gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f6ef6ac6-bf1f-41ed-bd3f-a3806a76a336",
		Name:            "Maze Rusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{dgmMulticoloredCreaturesHave("haste")},
	})
}
