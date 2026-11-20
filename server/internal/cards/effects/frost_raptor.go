package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frost Raptor — Snow Creature — Bird {2}{U}, 2/2 (Coldsnap, #5986):
//
//	"Flying
//	 {S}{S}: This creature gains shroud until end of turn. ({S} can be
//	 paid with one mana from a snow source. A creature with shroud can't
//	 be the target of spells or abilities.)"
//
// Advanced Hoverguard's self-grant at a snow price: each {S} is one mana
// from a snow source (CR 107.4h, #5986). Shroud is the keyword the
// targeting choke point already honours, so a spell already aimed at the
// Raptor is left with an illegal target when it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ca01d83-7794-4952-8cd8-c2ecb32cffd5",
		Name:            "Frost Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{S}{S}: This creature gains shroud until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{S}{S}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("shroud", "Frost Raptor — shroud"),
		}},
	})
}
