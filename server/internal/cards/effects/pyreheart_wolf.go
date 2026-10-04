package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyreheart Wolf — Creature — Wolf {2}{R}, 1/1:
//
//	"Whenever this creature attacks, creatures you control gain menace
//	 until end of turn.
//	 Undying"
//
// The creatures are chosen as the trigger resolves (CR 611.2c): the
// ones you control then, attacking or not. Undying is PrintedKeywords
// (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b97ccb69-e76c-4962-91ca-c7fd857140e8",
		Name:            "Pyreheart Wolf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pyreheart Wolf — creatures you control gain menace until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"menace"},
					Label:    "Pyreheart Wolf — menace until end of turn",
				})),
		},
	})
}
