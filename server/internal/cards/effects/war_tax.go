package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War Tax — Enchantment {2}{U}:
//
//	"{X}{U}: This turn, creatures can't attack unless their controller
//	 pays {X} for each attacking creature they control."
//
// Each resolution registers an any-defender ModAttackTax record (#3456,
// game/attack_tax_duration.go) whose mana is the activation's X, timed
// to end of turn. It taxes an attack on any player or planeswalker, and
// the records of several activations stack. X=0 taxes nothing.
//
// No simplification.
func init() {
	const name = "War Tax"
	Register(Spec{
		OracleID:     "fd5ff6ad-f3f1-470f-adef-fb4ce6536365",
		Name:         name,
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{X}{U}: This turn, creatures can't attack unless their controller pays {X} for each attacking creature they control.",
			Cost:    ManaCost("{X}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.XValue <= 0 {
					return nil
				}
				g.AnyAttackTaxForEffect(item.SourceCardID, fmt.Sprintf("{%d}", item.XValue),
					g.UntilEndOfTurnDuration(),
					fmt.Sprintf("%s — creatures can't attack unless their controller pays {%d} for each attacking creature they control", name, item.XValue))
				return nil
			},
		}},
	})
}
