package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Searing Spear Askari — Creature — Human Knight {2}{R}, 2/2 (Mirage,
// #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {1}{R}: This creature gains menace until end of turn. (It can't be
//	 blocked except by two or more creatures.)"
//
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a). Menace is a layer-6 grant pinned to the source through the
// shared thisGainsKeywordUntilEndOfTurn body (Cindering Cutthroat's), and
// the block legality check reads it (CR 702.111b): a lone blocker is
// refused. Each creature without flanking among two blockers gets -1/-1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d3c68573-95b1-4db5-bbea-363b86f713c1",
		Name:            "Searing Spear Askari",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{Flanking("Searing Spear Askari")},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"menace"}}},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("menace", "Searing Spear Askari — menace until end of turn"),
		}},
	})
}
