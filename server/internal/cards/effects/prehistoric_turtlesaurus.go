package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prehistoric Turtlesaurus — Creature — Mutant Ninja Turtle Dinosaur
// {4}{G}, 4/5 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"This spell costs {1} less to cast if you control a creature with a
//	 +1/+1 counter on it.
//	 Vigilance"
//
// The discount is a self cost modifier over the caster's own battlefield
// (Hamza's CreatureWithCounter, Venom's Hunger's YouControlA), judged when
// the total cost is determined (CR 601.2f) with the spell on the stack, so
// it never counts itself. The mana value stays 5 (the ruling). Vigilance
// rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e00aa7d7-3c11-463d-a7f5-8d001ea5c021",
		Name:            "Prehistoric Turtlesaurus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a creature with a +1/+1 counter on it.",
				YouControlA(CreatureWithCounter(game.CounterPlusOne))),
		},
	})
}
