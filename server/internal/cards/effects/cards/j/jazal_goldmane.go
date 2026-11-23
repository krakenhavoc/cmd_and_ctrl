package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jazal Goldmane — Legendary Creature — Cat Warrior {2}{W}{W}, 4/4
// (Foundations, #4465):
//
//	"First strike
//	 {3}{W}{W}: Attacking creatures you control get +X/+X until end of
//	 turn, where X is the number of attacking creatures."
//
// Marshal of the Lost's X (every attacking creature, whoever controls it,
// counted once as the ability resolves, CR 608.2h) on Tyvar, the
// Pummeler's team pump, narrowed to the attacking creatures you control
// then (CR 611.2c). Outside combat X is 0 and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "609d13dc-c850-4877-8ad3-e84171b58069",
		Name:            "Jazal Goldmane",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}{W}: Attacking creatures you control get +X/+X until end of turn, where X is the number of attacking creatures.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{W}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := len(effects.AttackingCreatures(g))
				if x <= 0 {
					return nil
				}
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.ControlledBy(item.Controller), effects.AttackingCreature()),
					Power:     x,
					Toughness: x,
					Label:     "Jazal Goldmane — attacking creatures you control get +X/+X",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
