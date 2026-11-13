package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Femeref Knight — Creature — Human Knight {2}{W}, 2/2 (Mirage, #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {W}: This creature gains vigilance until end of turn."
//
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a). Vigilance is a layer-6 grant pinned to the source through
// the shared thisGainsKeywordUntilEndOfTurn body; it matters only if the
// Knight has it as attackers are declared (CR 702.20b), so it is
// activated before combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4365521e-996d-4eca-bbc9-b0988fa973ae",
		Name:            "Femeref Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{RetiredFlankingRow("Femeref Knight")},
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature gains vigilance until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"vigilance"}}},
			Cost:    ManaCost("{W}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("vigilance", "Femeref Knight — vigilance until end of turn"),
		}},
	})
}
