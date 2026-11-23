package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Esquire of the King — Creature — Human Soldier {W}, 1/1 (The Hobbit):
//
//	"{4}{W}, {T}: Creatures you control get +1/+1 until end of turn.
//	 This ability costs {2} less to activate if you control a legendary
//	 creature."
//
// The discount is the ability's own cost clause (ActivatedAbility.
// CostModifiers, CR 601.2f via 602.2b), read as the cost is determined:
// {2} less of the generic part, once, however many legendary creatures
// you control, so {2}{W} with one. Only a creature counts: a legendary
// artifact or an opponent's legend does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ccaa78e-f3e5-4de4-b53c-cecd73a73b9c",
		Name:         "Esquire of the King",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}, {T}: Creatures you control get +1/+1 until end of turn. This ability costs {2} less to activate if you control a legendary creature.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{4}{W}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLess(2, "This ability costs {2} less to activate if you control a legendary creature.",
					effects.YouControlA(effects.And(effects.Creature(), effects.Legendary()))),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Esquire of the King — creatures you control get +1/+1"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
