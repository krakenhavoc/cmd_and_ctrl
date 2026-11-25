package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Chainwhirler — Creature — Goblin Warrior {R}{R}{R}, 3/3 (Dominaria,
// S82 OLD, #5530):
//
//	"First strike
//	 When this creature enters, it deals 1 damage to each opponent and each
//	 creature and planeswalker they control."
//
// One damage instance (CR 120.1) to every opponent and to every creature and
// planeswalker an opponent controls, read as the trigger resolves
// (DamageEachMatchingAndEachOpponent). Declared as a creature sweep that
// spares its controller's side.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0342b085-be20-475a-bc38-59b9cb6012e8",
		Name:            "Goblin Chainwhirler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisEnters("Goblin Chainwhirler — 1 damage to each opponent and each creature and planeswalker they control",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DamageEachMatchingAndEachOpponent(effects.NewContext(g, item), effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()), 1)
					}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}),
		},
	})
}
