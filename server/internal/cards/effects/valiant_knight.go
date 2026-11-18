package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b5c03abf-4297-4402-9107-539e92d41f2f",
		Name:         "Valiant Knight",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}{W}: Knights you control gain double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{W}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{Match: And(Creature(), YouControl(), Subtype("Knight")), Keywords: []string{"double strike"},
					Label: "Valiant Knight — Knights you control gain double strike"}.Apply(NewContext(g, item))
			},
		}},
	})
}
