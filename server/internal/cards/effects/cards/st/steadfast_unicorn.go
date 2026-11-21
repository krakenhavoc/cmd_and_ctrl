package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfast Unicorn — {W} Creature — Unicorn 1/2 (Foundations Jumpstart):
//
//	"{3}{W}: Creatures you control get +1/+1 and gain vigilance until end
//	 of turn. Activate only during your turn."
//
// The set is the creatures you control as the ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c823d38c-e73c-4e09-ab20-c5e7de1f1c64",
		Name:         "Steadfast Unicorn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}{W}: Creatures you control get +1/+1 and gain vigilance until end of turn. Activate only during your turn.",
			Purpose:   game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1, Keywords: []string{"vigilance"}}, Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{3}{W}"),
			Condition: effects.DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				match := effects.And(effects.Creature(), effects.YouControl())
				if err := (effects.BoostUntilEOT{Match: match, Power: 1, Toughness: 1,
					Label: "Steadfast Unicorn — creatures you control get +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"vigilance"},
					Label: "Steadfast Unicorn — creatures you control gain vigilance"}.Apply(ctx)
			},
		}},
	})
}
