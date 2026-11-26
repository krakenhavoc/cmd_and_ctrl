package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paragon of the Amesha — Creature — Human Knight {2}{W}, 2/2 (Conflux):
//
//	"First strike
//	 {W}{U}{B}{R}{G}: Until end of turn, this creature becomes an Angel, gets
//	 +3/+3, and gains flying and lifelink."
//
// One until-end-of-turn record over the Paragon itself (CR 611.2c, only while
// it is still the permanent that activated, CR 400.7): "becomes an Angel"
// replaces its creature types (layer 4, CR 205.1a; it is no longer a Human
// Knight), +3/+3 is a layer-7 change, and flying and lifelink are layer-6
// grants beside the printed first strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4e4991b3-11c4-4408-a752-67e51f0fae44",
		Name:            "Paragon of the Amesha",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: Until end of turn, this creature becomes an Angel, gets +3/+3, and gains flying and lifelink.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.UntilEndOfTurn(ctx, item.SourceCardID, nil,
					"Paragon of the Amesha — becomes an Angel, gets +3/+3 and gains flying and lifelink",
					game.SetCreatureTypesMod(ctx.TextSubtype("Angel")),
					game.ModifyPTMod(3, 3),
					game.AddKeywordsMod("flying", "lifelink"))
			},
		}},
	})
}
