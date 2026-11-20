package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maze Behemoth — Creature — Elemental {5}{G}, 5/4:
//
//	"Trample
//	 Multicolored creatures you control have trample."
//
// The printed trample plus a layer-6 grant (CR 613.1f) of trample to every
// creature its controller controls that is two or more colours
// (CR 105.2b), read through the layers like Rienne's multicolored anthem.
// A monocolored or colourless creature gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d54c258f-1914-449a-857d-11f9478bff63",
		Name:            "Maze Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{dgmMulticoloredCreaturesHave("trample")},
	})
}
