package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War-Spike Changeling — Creature — Shapeshifter {3}{R}, 3/3 (Morningtide,
// #5592):
//
//	"Changeling (This card is every creature type.)
//	 {R}: This creature gains first strike until end of turn."
//
// Changeling is every creature type in every zone (CR 702.73a). The
// activation is Burning Shield Askari's first strike, for {R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b6ddb093-1fff-4648-b7a1-141a3875599c",
		Name:            "War-Spike Changeling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "War-Spike Changeling — first strike until end of turn"),
		}},
	})
}
