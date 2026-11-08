package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sun Warriors — Creature — Human Warrior Ally {2}{R}{W}, 3/5 (Avatar: The
// Last Airbender, #4459):
//
//	"Firebending X, where X is the number of creatures you control.
//	 (Whenever this creature attacks, add X {R}. This mana lasts until end
//	 of combat.)
//	 {5}: Create a 1/1 white Ally creature token."
//
// Firebending is Avatar Roku's (`firebending`) with its count read as the
// trigger resolves: firebending is not a mana ability and uses the stack
// (the ruling), so X is the creatures its controller controls then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "513a2fca-fea4-434e-9b25-b0c0d7f322b2",
		Name:         "Sun Warriors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sun Warriors — firebending X", func(g *game.Game, item *game.StackItem) error {
				// X of zero adds nothing but still firebends (#5276).
				return firebending(b04CreaturesControlled(g, item.Controller)).Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}: Create a 1/1 white Ally creature token.",
			Cost:    ManaCost("{5}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Ally"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
