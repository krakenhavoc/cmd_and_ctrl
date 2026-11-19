package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sprite Noble — {1}{U}{U} Creature — Faerie Noble 2/2 (Foundations
// Jumpstart):
//
//	"Flying
//	 Other creatures you control with flying get +0/+1.
//	 {T}: Other creatures you control with flying get +1/+0 until end of turn."
//
// The static is a layer 7c modify read on every recompute
// (b29OtherCreaturesYouControlWithFlying). The activated ability locks its
// set as it resolves (CR 611.2c): only the creatures that have flying then
// get the bonus.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bf3e3004-4133-4542-9ee3-2fb98e759ffa",
		Name:            "Sprite Noble",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{b16Anthem(b29OtherCreaturesYouControlWithFlying, 0, 1)},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Other creatures you control with flying get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl(), OtherThan(item.SourceCardID), HasKeyword("flying")),
					Power: 1,
					Label: "Sprite Noble — other creatures you control with flying get +1/+0",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
