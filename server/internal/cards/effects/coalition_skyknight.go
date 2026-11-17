package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coalition Skyknight — Creature — Human Knight {3}{W}, 2/2 (Dominaria
// United, #5487, seam #5658):
//
//	"Flying
//	 Enlist (As this creature attacks, you may tap a nonattacking creature
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
		OracleID:        "dc763f20-73e7-4db7-9e98-6f419b8fba84",
		Name:            "Coalition Skyknight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordEnlist},
	})
}
