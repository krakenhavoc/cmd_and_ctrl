package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandsteppe Mastodon — Creature — Elephant {5}{G}{G}, 5/5 (Fate Reforged,
// #5552):
//
//	"Reach
//	 When this creature enters, bolster 5. (Choose a creature with the
//	 least toughness among creatures you control and put five +1/+1
//	 counters on it.)"
//
// Reach is printed. Bolster (CR 701.39a) chooses as the trigger resolves, so
// the Mastodon itself is a candidate; ties are its controller's pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99fa202b-052c-4627-b78c-f0b323f7bdf3",
		Name:            "Sandsteppe Mastodon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sandsteppe Mastodon — bolster 5", Do(Bolster{N: 5})),
		},
	})
}
