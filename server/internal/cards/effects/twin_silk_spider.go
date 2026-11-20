package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twin-Silk Spider — Creature — Spider {2}{G}, 1/2 (Modern Horizons, #5519):
//
//	"Reach
//	 When this creature enters, create a 1/2 green Spider creature token
//	 with reach."
//
// Brood Weaver's token on an enters trigger instead of a dies one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2193302-402a-4cf0-8014-5facecad647b",
		Name:            "Twin-Silk Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Twin-Silk Spider — create a 1/2 green Spider creature token with reach",
				Do(CreateToken{Template: TokenCard("1/2 green Spider with reach"), N: 1})),
		},
	})
}
