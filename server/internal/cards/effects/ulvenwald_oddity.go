package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ulvenwald Oddity // Ulvenwald Behemoth — {2}{G}{G} Creature — Beast 4/4 //
// Creature — Beast Horror 8/8 (Innistrad: Crimson Vow):
//
//	Front: "Trample, haste
//	        {5}{G}{G}: Transform this creature."
//	Back:  "Trample, haste
//	        Other creatures you control get +1/+1 and have trample and haste."
//
// The Behemoth's anthem is "other creatures you control": the Behemoth is
// untouched and an opponent's creatures are not pumped.
//
// No simplification.
func init() {
	const oracle = "90379825-1894-4b79-bf72-445175e97ffb"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Ulvenwald Oddity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}{G}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{5}{G}{G}"),
			Effect:  Do(TransformThis{}),
		}},
	})
	others := TribeFilter{Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Ulvenwald Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Static: []game.StaticAbility{
			TribalAnthem(others, 1, 1),
			TribalKeywordGrant(others, "trample"),
			TribalKeywordGrant(others, "haste"),
		},
	})
}
