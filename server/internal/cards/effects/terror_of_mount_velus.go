package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terror of Mount Velus — Creature — Dragon {5}{R}{R}, 5/5
// (Foundations, #4465):
//
//	"Flying
//	 Double strike
//	 When this creature enters, creatures you control gain double strike
//	 until end of turn."
//
// Two printed keywords, and War Machine's GrantKeywordUntilEOT over a
// Match: the creatures you control are locked in as the trigger
// resolves (CR 611.2c), so a creature that enters later that turn
// doesn't gain double strike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8161248c-16a3-47db-a3ae-cdfd878d49e0",
		Name:            "Terror of Mount Velus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Terror of Mount Velus — creatures you control gain double strike until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"double strike"},
					Label:    "Terror of Mount Velus — double strike",
				})),
		},
	})
}
