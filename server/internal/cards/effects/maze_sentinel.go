package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maze Sentinel — Creature — Elemental {5}{W}, 3/6:
//
//	"Vigilance
//	 Multicolored creatures you control have vigilance."
//
// The printed vigilance plus a layer-6 grant (CR 613.1f) of vigilance to
// every creature its controller controls that is two or more colours
// (CR 105.2b), read through the layers like Rienne's multicolored anthem.
// A monocolored or colourless creature gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "635360cd-0114-47d1-8a98-e6a3678652a8",
		Name:            "Maze Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{dgmMulticoloredCreaturesHave("vigilance")},
	})
}
