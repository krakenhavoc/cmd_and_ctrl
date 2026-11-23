package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Ulvenwald Oddity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}{G}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{G}{G}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	others := effects.TribeFilter{Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Ulvenwald Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(others, 1, 1),
			effects.TribalKeywordGrant(others, "trample"),
			effects.TribalKeywordGrant(others, "haste"),
		},
	})
}
