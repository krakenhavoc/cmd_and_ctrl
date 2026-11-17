package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coalition Warbrute — Creature — Minotaur Berserker {3}{R}, 3/4
// (Dominaria United, #5487, seam #5658):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)
//	 Trample"
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
// The added power tramples over with the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "84d67d87-2696-4c2f-9f10-29d5832bfbb6",
		Name:            "Coalition Warbrute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist, "trample"},
	})
}
