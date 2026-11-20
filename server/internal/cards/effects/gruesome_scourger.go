package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gruesome Scourger — Creature — Orc Warrior {3}{B}{B}, 3/3 (Core Set 2020, #5518):
//
//	"When this creature enters, it deals damage to target opponent or
//	 planeswalker equal to the number of creatures you control."
//
// The count is read as the trigger resolves, the Scourger included if it
// is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b2993fa-b212-4eaa-b3e8-f79c2f2e4cbe",
		Name:         "Gruesome Scourger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Gruesome Scourger — damage to target opponent or planeswalker equal to the number of creatures you control",
					func(g *game.Game, item *game.StackItem) error {
						return DealDamageToTheTarget(creaturesControlledBy(g, item.Controller))(g, item)
					}),
				targetOpponentOrPlaneswalker(),
			),
		},
	})
}
