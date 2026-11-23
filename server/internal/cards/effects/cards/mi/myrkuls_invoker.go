package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myrkul's Invoker — {2}{B} Creature — Cat Rogue 2/3
//
//	"Psychic Blades — {8}: Creatures you control get +2/+0 and gain menace
//	 until end of turn."
//
// The creatures you control as it resolves are the ones affected (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bc9f027-8a2a-4447-8458-66ed40cb5b0d",
		Name:         "Myrkul's Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Psychic Blades — {8}: Creatures you control get +2/+0 and gain menace until end of turn.",
			Cost:    effects.ManaCost("{8}"),
			Purpose: game.Purpose{Answers: game.AnswerPump, TeamPump: &game.Pump{Power: 2, Keywords: []string{"menace"}}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				mine := effects.And(effects.Creature(), effects.YouControl())
				if err := (effects.BoostUntilEOT{Match: mine, Power: 2, Label: "Myrkul's Invoker — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"menace"}, Label: "Myrkul's Invoker — menace"}.Apply(ctx)
			},
		}},
	})
}
