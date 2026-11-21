package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valiant Knight — Creature — Human Knight {3}{W}, 3/4:
//
//	"Other Knights you control get +1/+1.
//	 {3}{W}{W}: Knights you control gain double strike until end of turn."
//
// Marshal of Zhalfir's anthem; the activated ability is The Howling
// Commandos' keyword grant over a tribe, locked as it resolves (CR 611.2c).
// The Knight itself is included in "Knights you control".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5c03abf-4297-4402-9107-539e92d41f2f",
		Name:         "Valiant Knight",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}{W}: Knights you control gain double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{W}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Knight")), Keywords: []string{"double strike"},
					Label: "Valiant Knight — Knights you control gain double strike"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
