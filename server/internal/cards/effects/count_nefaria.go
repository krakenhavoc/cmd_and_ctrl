package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Count Nefaria — Legendary Creature — Human Villain {6}{B}, 6/6 (Marvel
// Super Heroes Commander):
//
//	"This spell costs {3} less to cast if you've sacrificed a permanent
//	 this turn.
//	 Flying"
//
// A conditional self cost modifier (Mortality Spear's shape) reading the
// turn tally's permanents sacrificed for the caster, so a Treasure or a
// Clue sacrificed this turn counts and one an opponent sacrificed does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ec4f81e-ede4-4d45-b768-16583f6f5b8d",
		Name:            "Count Nefaria",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if you've sacrificed a permanent this turn.",
				func(q game.CostQuery) bool {
					return q.Game != nil && q.Game.TurnTallyFor(q.Controller).PermanentsSacrificed > 0
				}),
		},
	})
}
