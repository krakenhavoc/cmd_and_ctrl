package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonsoul Knight — Creature — Human Knight {2}{R}, 2/2 (Conflux):
//
//	"First strike
//	 {W}{U}{B}{R}{G}: Until end of turn, this creature becomes a Dragon, gets
//	 +5/+3, and gains flying and trample."
//
// One until-end-of-turn record over the Knight itself (CR 611.2c, only while
// it is still the permanent that activated, CR 400.7): "becomes a Dragon"
// replaces its creature types (layer 4, CR 205.1a; it is no longer a Human
// Knight), +5/+3 is a layer-7 change, and flying and trample are layer-6
// grants beside the printed first strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8fb5a40-4e32-4817-bf56-484add976475",
		Name:            "Dragonsoul Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: Until end of turn, this creature becomes a Dragon, gets +5/+3, and gains flying and trample.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.UntilEndOfTurn(ctx, item.SourceCardID, nil,
					"Dragonsoul Knight — becomes a Dragon, gets +5/+3 and gains flying and trample",
					game.SetCreatureTypesMod(ctx.TextSubtype("Dragon")),
					game.ModifyPTMod(5, 3),
					game.AddKeywordsMod("flying", "trample"))
			},
		}},
	})
}
