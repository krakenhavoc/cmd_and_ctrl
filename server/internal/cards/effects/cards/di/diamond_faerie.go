package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diamond Faerie — Snow Creature — Faerie {2}{G}{W}{U}, 3/3 (Coldsnap, S82 OLD b560):
//
//	"Flying
//	 {1}{S}: Snow creatures you control get +1/+1 until end of turn."
//
// The snow creatures you control as the ability resolves are the ones pumped
// (CR 611.2c), the Faerie among them; {S} is paid with snow mana only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a25ccf3-0fe1-447c-bdf4-57ec31b19a0f",
		Name:            "Diamond Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{S}: Snow creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{1}{S}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.Snow()), Power: 1, Toughness: 1,
					Label: "Diamond Faerie — snow creatures get +1/+1 until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
