package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "8161248c-16a3-47db-a3ae-cdfd878d49e0",
		Name:            "Terror of Mount Velus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Terror of Mount Velus — creatures you control gain double strike until end of turn",
				effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"double strike"},
					Label:    "Terror of Mount Velus — double strike",
				})),
		},
	})
}
