package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benalish Faithbonder — Creature — Human Cleric {1}{W}, 1/3 (Dominaria
// United, #5487, seam #5658):
//
//	"Vigilance
//	 Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)"
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
// Vigilance keeps the Faithbonder untapped as it attacks (CR 702.20b); it
// does nothing for the creature it enlists, which is tapped as a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "278db664-4235-40f0-a926-3074a0e038c1",
		Name:            "Benalish Faithbonder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", game.KeywordEnlist},
	})
}
