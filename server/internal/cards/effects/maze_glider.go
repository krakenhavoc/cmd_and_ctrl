package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maze Glider — Creature — Elemental {5}{U}, 3/5:
//
//	"Flying
//	 Multicolored creatures you control have flying."
//
// The printed flying plus a layer-6 grant (CR 613.1f) of flying to every
// creature its controller controls that is two or more colours
// (CR 105.2b), read through the layers like Rienne's multicolored anthem.
// A monocolored or colourless creature gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2388a064-ae84-4eac-b1d9-819a40052d12",
		Name:            "Maze Glider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{dgmMulticoloredCreaturesHave("flying")},
	})
}
