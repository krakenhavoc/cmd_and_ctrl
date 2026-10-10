package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightscape Familiar — Creature — Zombie {1}{B}, 1/1:
//
//	"Blue spells and red spells you cast cost {1} less to cast.
//	 {1}{B}: Regenerate this creature."
//
// The discount is a two-colour Or, Goblin Anarchomancer's shape
// (nightscapeBlueOrRedSpell, local to this card for the same reason
// redOrGreenSpell is local to that one). The regeneration ability is
// the ordinary CR 701.19 shield (game/regeneration.go) over the
// source itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57296ea3-3c0d-49b7-bc08-0d0d8414e9ad",
		Name:         "Nightscape Familiar",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Blue spells and red spells you cast cost {1} less to cast.",
				YourSpell(), nightscapeBlueOrRedSpell()),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{1}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}

func nightscapeBlueOrRedSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.HasColor("U") || q.Card.HasColor("R") }
}
