package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hamza, Guardian of Arashin — Legendary Creature — Elephant Warrior
// {4}{G}{W}, 5/5:
//
//	"This spell costs {1} less to cast for each creature you control
//	 with a +1/+1 counter on it.
//	 Creature spells you cast cost {1} less to cast for each creature
//	 you control with a +1/+1 counter on it."
//
// #746: the same count in both slots — the spell's own reduction, and
// the battlefield static for later creature spells.
func init() {
	withCounter := CreatureWithCounter("+1/+1")
	Register(Spec{
		OracleID:     "0cde6a29-d597-4feb-8b2b-528d2d57a1f3",
		Name:         "Hamza, Guardian of Arashin",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsYouControl(withCounter),
				"This spell costs {1} less to cast for each creature you control with a +1/+1 counter on it."),
		},
		CostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsYouControl(withCounter),
				"Creature spells you cast cost {1} less to cast for each creature you control with a +1/+1 counter on it.",
				YourSpell(), CreatureSpell()),
		},
	})
}
