package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leering Gargoyle — Creature — Gargoyle {1}{W}{U}, 2/2 (Mirage, S82 OLD b797,
// #5633):
//
//	"Flying
//	 {T}: This creature gets -2/+2 and loses flying until end of turn."
//
// A tap ability on the source itself: the shrink is the shared this-gets body
// and the loss of flying is the layer-6 ability-removing effect with its own
// timestamp (CR 613.1f), both ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc0907eb-a3fc-40ef-950d-b05aee9d5da7",
		Name:            "Leering Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature gets -2/+2 and loses flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(-2, 2, "Leering Gargoyle — -2/+2 until end of turn")(g, item); err != nil {
					return err
				}
				return effects.ThisLosesKeywordUntilEndOfTurn("flying", "Leering Gargoyle — loses flying until end of turn")(g, item)
			},
		}},
	})
}
