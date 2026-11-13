package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage Owl — Creature — Bird {1}{U}, 1/1 (Weatherlight, #5629):
//
//	"Flying
//	 When this creature enters, look at the top four cards of your library,
//	 then put them back in any order."
//
// Flying is the printed keyword. The enters trigger is the LookAtTop
// prompt (Ponder's, Discombobulate's four): every card goes back on top,
// in the order the controller picks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d47fc902-51f7-4ad1-8ee4-c973e32192b8",
		Name:            "Sage Owl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sage Owl — look at the top four cards of your library, then put them back in any order", Do(LookAtTop{N: 4})),
		},
	})
}
