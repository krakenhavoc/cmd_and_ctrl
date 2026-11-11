package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wickerbough Elder — Creature — Treefolk Shaman {3}{G}, 4/4 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"This creature enters with a -1/-1 counter on it.
//	 {G}, Remove a -1/-1 counter from this creature: Destroy target
//	 artifact or enchantment."
//
// The counter comes in by an entry replacement; removing it is a cost paid
// at announce, so the ability can be used once unless another counter is
// added.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d794629-9014-460a-8d3b-785eadc1cdb5",
		Name:         "Wickerbough Elder",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 1, "Wickerbough Elder: enters with a -1/-1 counter"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{G}, Remove a -1/-1 counter from this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{G}"), RemoveCountersFromThis(game.CounterMinusOne, 1)),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return destroyTheTargetPermanent(item, NewContext(g, item))
			},
		}},
	})
}
