package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerie Bowmasters — 3/4 Dog Archer for {2}{G}{G}:
//
//	"Reach. Megamorph {5}{G}"
//
// Megamorph is morph plus one clause — "put a +1/+1 counter on it"
// when the permanent is turned face up (CR 702.37b) — and that clause
// rides the keyword rather than this file: the counter is placed by
// the turn-face-up special action, before the CR 708.8 event, so
// anything watching the permanent turn over already sees the 4/5.
//
// In the catalog for the counter. Ainok Tracker next door proves plain
// morph; what this card proves is that the two differ by a declaration
// and not by a code path.
func init() {
	Register(Spec{
		OracleID:        "43ace4d5-9006-4bad-bd17-3d368c20564d",
		Name:            "Aerie Bowmasters",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{
			Megamorph("{5}{G}"),
		},
	})
}
