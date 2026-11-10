package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shore Lurker — Creature — Merfolk Scout {3}{W}, 3/3 (Lorwyn Eclipsed,
// #4458):
//
//	"Flying
//	 When this creature enters, surveil 1."
//
// Flying is printed data; the surveil is A.I.M. Synthoids's with a count of
// one (a stack ability, so an opponent can respond).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4288ba56-ddd0-4bd9-ba61-daf03e9f2d7f",
		Name:            "Shore Lurker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shore Lurker — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
