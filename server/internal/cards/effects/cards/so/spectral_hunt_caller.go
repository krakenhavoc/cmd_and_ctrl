package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Hunt-Caller — Creature — Wolf Spirit {4}{G}, 4/4 (Jumpstart 2022,
// #5948):
//
//	"{5}{G}: Creatures you control get +1/+1 and gain trample until end of
//	 turn."
//
// The set of creatures is locked as the ability resolves (CR 611.2c), as
// Cabaretti Charm's pump is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d03ef35b-0609-46b5-ab8a-64d8007eab26",
		Name:         "Spectral Hunt-Caller",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: Creatures you control get +1/+1 and gain trample until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1, Keywords: []string{"trample"}}, Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				mine := effects.And(effects.Creature(), effects.YouControl())
				if err := (effects.BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Spectral Hunt-Caller — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Spectral Hunt-Caller — trample"}.Apply(ctx)
			},
		}},
	})
}
