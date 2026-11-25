package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Encampment Keeper — Creature — Dog {W}, 1/1 (Ixalan):
//
//	"First strike
//	 {7}{W}, {T}, Sacrifice this creature: Creatures you control get +2/+2
//	 until end of turn."
//
// Ixalli's Keeper's cost shape. The tap and the sacrifice are costs, so it
// can't be activated the turn it arrives (CR 302.6). The creatures it affects
// are the ones its controller controls as the ability resolves (CR 611.2c),
// the Keeper already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86f2890d-3e5a-48db-bb22-a6cb8cbd39ce",
		Name:            "Encampment Keeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{W}, {T}, Sacrifice this creature: Creatures you control get +2/+2 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 2, Toughness: 2}, Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{7}{W}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2,
					Label: "Encampment Keeper — creatures you control get +2/+2",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
