package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Researcher — Creature — Goblin Wizard {3}{R}, 3/3:
//
//	"When this creature enters, exile the top card of your library.
//	 During any turn you attacked with this creature, you may play that
//	 card."
//
// A conditional exile play (#3192, ADR 0066's second 2026-10-13
// amendment). The permission lasts for as long as the card stays
// exiled and is live only during a turn its holder attacked with THIS
// creature: game.CondAttackedWithObject names the Researcher's object
// (instance and CR 400.7 epoch) as the trigger went on the stack, so a
// Researcher that was flickered or died, or another Goblin Researcher,
// does not open it. The card is dormant, not lost, on every turn it did
// not attack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8762dfa-a027-4e1c-a2fa-bfdab1de102a",
		Name:         "Goblin Researcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Goblin Researcher — exile the top card of your library", func(g *game.Game, item *game.StackItem) error {
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
					WhileExiled: true,
					Condition:   attackedWithObject(item),
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
