package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hand of Justice — Creature — Avatar {5}{W}, 2/6 (Fallen Empires, S82 OLD b854,
// #5637):
//
//	"{T}, Tap three untapped white creatures you control: Destroy target
//	 creature."
//
// The three creatures are a cost beside the Hand's own {T}, so the Hand can't
// be one of them (CR 118.3); they are not the {T} symbol, so creatures that
// arrived this turn may pay (CR 302.6), and the cost does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af8eb20a-dfa2-49a4-892b-26d7e276fcd2",
		Name:         "Hand of Justice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Tap three untapped white creatures you control: Destroy target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.TapCost(), effects.S82OldB854TapNOthers(3, "three untapped white creatures you control", effects.Creature(), effects.OfColor("W"))),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
			},
		}},
	})
}
