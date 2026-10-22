package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnarled Professor — Creature — Treefolk Druid {2}{G}{G}, 5/4 (EDHREC
// rank 24225):
//
//	"Trample
//	 When this creature enters, learn. (You may reveal a Lesson card you
//	 own from outside the game and put it into your hand, or discard a
//	 card to draw a card.)"
//
// An enters trigger whose instruction is learn (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e66083f0-62a9-4ff7-8bb0-3eafc6d2e95d",
		Name:            "Gnarled Professor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gnarled Professor — learn", Do(Learn{})),
		},
	})
}
