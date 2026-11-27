package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Dawn — Creature — Human Knight {1}{W}{W}, 2/2 (Tempest, S82 OLD
// b753, #5628):
//
//	"First strike
//	 {W}{W}: This creature gains protection from the color of your choice
//	 until end of turn."
//
// Cartel Aristocrat's grant: the colour is chosen as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51262155-f1be-45a8-a21f-94e920a93e55",
		Name:            "Knight of Dawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}: This creature gains protection from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{W}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				effects.ChooseProtectionColorFor(ctx, ctx.Controller(), item.SourceCardID, "Knight of Dawn",
					"Knight of Dawn — choose a color to gain protection from")
				return nil
			},
		}},
	})
}
