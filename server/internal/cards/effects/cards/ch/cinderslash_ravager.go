package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinderslash Ravager — Creature — Phyrexian Warrior {4}{R}{G}, 5/5
// (Phyrexia: All Will Be One):
//
//	"This spell costs {1} less to cast for each permanent you control
//	 with oil counters on it.
//	 Vigilance
//	 When this creature enters, it deals 1 damage to each creature your
//	 opponents control."
//
// The reduction is a self cost modifier counting your permanents with an
// oil counter (generic mana only, CR 601.2f). The damage is a one-damage
// sweep of every creature an opponent controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a1c6bd47-19ef-44f9-921a-e4ffa80b25e1",
		Name:            "Cinderslash Ravager",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.PermanentsOnBattlefield(effects.And(effects.YouControl(), effects.S82HasOilCounter)),
				"This spell costs {1} less to cast for each permanent you control with oil counters on it."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cinderslash Ravager — it deals 1 damage to each creature your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.OpponentControls()), 1)
				}),
		},
	})
}
