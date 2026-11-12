package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barkweave Crusher — Creature — Elemental Warrior {3}{G}, 2/5 (Dominaria
// United, #5487, seam #5658):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)"
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cbac75c6-6006-4025-b893-8828965b5f97",
		Name:            "Barkweave Crusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
	})
}
