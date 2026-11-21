package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flexible Waterbender — Creature — Human Warrior Ally {3}{U}, 2/5
// (Avatar: The Last Airbender, #4459):
//
//	"Vigilance
//	 Waterbend {3}: This creature has base power and toughness 5/2 until
//	 end of turn. (While paying a waterbend cost, you can tap your
//	 artifacts and creatures to help. Each one pays for {1}.)"
//
// WaterbendCost("{3}") is the activated ability's whole cost (CR
// 701.67a), and the effect sets base power and toughness in layer 7b,
// Stature, Young Avenger's shape, so counters and pumps still apply on
// top. The ability does nothing to a new object if this one has left.
//
// No simplification.
func init() {
	const label = "Flexible Waterbender — base power and toughness 5/2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "8cf88fa8-2f47-4125-a59d-9c2b2d4765a7",
		Name:            "Flexible Waterbender",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Waterbend {3}: This creature has base power and toughness 5/2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.WaterbendCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil, label,
					game.SetBasePowerMod(5), game.SetBaseToughnessMod(2))
			},
		}},
	})
}
