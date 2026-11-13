package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zhalfirin Knight — Creature — Human Knight {2}{W}, 2/2 (Mirage, #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {W}{W}: This creature gains first strike until end of turn."
//
// Burning Shield Askari's card in white: flanking.go's trigger with the
// declared keyword (#5757, CR 702.25a), and the first-strike grant
// through the shared thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a534f06e-dae4-4d1a-96e9-7693bbb21e12",
		Name:            "Zhalfirin Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{Flanking("Zhalfirin Knight")},
		Activated: []ActivatedAbility{{
			Label:   "{W}{W}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    ManaCost("{W}{W}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "Zhalfirin Knight — first strike until end of turn"),
		}},
	})
}
