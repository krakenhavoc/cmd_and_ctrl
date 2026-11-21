package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Matrix — Artifact {4} (Mystery Booster 2):
//
//	"{T}: Target creature gets +1/+1 and gains flying, first strike, and
//	 trample until end of turn."
//
// One locked-in effect (CR 611.2c): the +1/+1 (layer 7c) and the three
// keywords (layer 6) ride one until-end-of-turn record.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4355507e-e3a4-4c49-a1f7-344150213a2f",
		Name:         "Power Matrix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Target creature gets +1/+1 and gains flying, first strike, and trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(game.TargetPurpose{
				Slot: 0, UntilEndOfTurn: game.TargetBenefit,
				Pump: &game.Pump{Power: 1, Toughness: 1, Keywords: []string{"flying", "first strike", "trample"}},
			})},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, ts[0].ID, nil, "Power Matrix — +1/+1, flying, first strike, trample",
					game.ModifyPTMod(1, 1), game.AddKeywordsMod("flying", "first strike", "trample"))
			},
		}},
	})
}
