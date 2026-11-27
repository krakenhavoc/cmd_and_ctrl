package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adarkar Windform — Snow Creature — Illusion {4}{U}, 3/3 (Coldsnap):
//
//	"Flying
//	 {1}{S}: Target creature loses flying until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd66b8ab-14a9-455b-98f6-5356df72a522",
		Name:            "Adarkar Windform",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{S}: Target creature loses flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.ManaCost("{1}{S}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, id, nil, "Adarkar Windform — loses flying until end of turn", game.RemoveKeywordsMod("flying"))
			},
		}},
	})
}
