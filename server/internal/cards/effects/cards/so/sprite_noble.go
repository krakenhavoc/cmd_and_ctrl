package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "bf3e3004-4133-4542-9ee3-2fb98e759ffa",
		Name:            "Sprite Noble",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.B16Anthem(effects.B29OtherCreaturesYouControlWithFlying, 0, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Other creatures you control with flying get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID), effects.HasKeyword("flying")),
					Power: 1,
					Label: "Sprite Noble — other creatures you control with flying get +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
