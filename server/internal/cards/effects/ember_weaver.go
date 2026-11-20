package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ember Weaver — Creature — Spider {2}{G}, 2/3 (Conflux, #5587):
//
//	"Reach (This creature can block creatures with flying.)
//	 As long as you control a red permanent, this creature gets +1/+0 and
//	 has first strike."
//
// Reach is a printed keyword (CR 702.17b). The conditional half is two
// statics with one condition (conB1WhileYouControlA): +1/+0 in layer 7c
// (CR 613.4c) and first strike in layer 6 (CR 613.1f).
//
// No simplification.
func init() {
	redPump := game.StaticAbility{
		Label:     "As long as you control a red permanent, this creature gets +1/+0.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: conB1WhileYouControlA("R"),
		Apply:     pumpBy(1, 0),
	}
	firstStrike := KeywordGrant(conB1WhileYouControlA("R"), "first strike")
	firstStrike.Label = "As long as you control a red permanent, this creature has first strike."
	Register(Spec{
		OracleID:        "c92221f7-76e8-483a-9d68-b19ff4d0baaf",
		Name:            "Ember Weaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static:          []game.StaticAbility{redPump, firstStrike},
	})
}
