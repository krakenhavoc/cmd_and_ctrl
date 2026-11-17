package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Monitor — Creature — Phyrexian Skeleton {3}{B}, 2/2 (Urza's
// Destiny, #5621):
//
//	"{B}: Regenerate this creature."
//
// Mortivore's regeneration row on a plain 2/2: each activation sets up
// one shield for this turn, used up the next time the Skeleton would be
// destroyed, which taps it, removes it from combat and heals its damage
// instead. No {T} in the cost, so it can be activated as often as you
// can pay, summoning sick or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1eb6c28d-ebac-4016-ad07-c90081ee4983",
		Name:         "Phyrexian Monitor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
