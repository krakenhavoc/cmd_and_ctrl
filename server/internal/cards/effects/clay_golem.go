package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clay Golem — Artifact Creature — Golem {4}, 4/4 (Forgotten Realms
// Commander, #5682):
//
//	"{6}, Roll a d8: Monstrosity X, where X is the result. (If this
//	 creature isn't monstrous, put X +1/+1 counters on it and it becomes
//	 monstrous.)
//	 Berserk — When this creature becomes monstrous, destroy target
//	 permanent."
//
// The roll is part of the cost (RollDieCost): the d8 is rolled as the
// ability is activated, after the {6} is paid (CR 601.2h), so the table
// knows X while the ability is on the stack and a second activation in
// response rolls again and pays again for nothing. The result is on the
// ability's payment record, and the monstrosity reads it as the ability
// resolves. A countered ability has still rolled its die.
//
// The Berserk trigger is mandatory and its target is any permanent, the
// Golem and its controller's own included; with nothing else on the
// battlefield it has to take the Golem.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8773c2ea-ef41-4403-bb52-07f0d70d9bbe",
		Name:         "Clay Golem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{MonstrosityXWhere(
			Plus(ManaCost("{6}"), RollDieCost(8)),
			"{6}, Roll a d8: Monstrosity X, where X is the result.",
			func(_ *game.Game, item *game.StackItem) int { return CostRollOf(item) },
		)},
		Triggered: []game.TriggeredAbility{clayGolemBerserk()},
	})
}

func clayGolemBerserk() game.TriggeredAbility {
	t := WhenBecomesMonstrous("Clay Golem — Berserk: destroy target permanent", destroyChosenPermanent)
	t.Targets = TargetPermanent("target permanent", Permanent())
	return t
}
