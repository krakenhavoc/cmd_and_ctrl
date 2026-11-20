package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mothdust Changeling — Creature — Shapeshifter {B}, 1/1 (Morningtide,
// #5592):
//
//	"Changeling (This card is every creature type.)
//	 Tap an untapped creature you control: This creature gains flying until
//	 end of turn."
//
// Changeling is every creature type in every zone (CR 702.73a). The cost
// is Captain Janeway's "tap an untapped creature you control": not the {T}
// symbol, so a creature that arrived this turn may pay it (CR 302.6), and
// the Changeling may tap itself to pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "52237e6b-19f6-405c-a668-545c4ccb1b6d",
		Name:            "Mothdust Changeling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated: []ActivatedAbility{{
			Label:   "Tap an untapped creature you control: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"flying"}}},
			Cost:    TapAnUntapped("an untapped creature you control", Creature()),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Mothdust Changeling — flying until end of turn"),
		}},
	})
}
