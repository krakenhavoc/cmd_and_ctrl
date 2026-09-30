package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Gatekeeper — Artifact Creature — Golem {2}{R}, 2/3:
//
//	"Whenever another creature you control enters, this creature deals
//	 1 damage to each opponent.
//	 Unearth {R} ({R}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next
//	 end step or if it would leave the battlefield. Unearth only as a
//	 sorcery.)"
//
// Impact Tremors' shape with "another" instead of "a" — the Gatekeeper
// itself entering does not ping anybody, including its own unearthed
// return. Unearth is the constructor: the return, the haste, the
// end-step exile and the leaves-the-battlefield redirect all come with
// the keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90fdfba8-f29e-44f9-91d2-7bf3c458a9c1",
		Name:         "Molten Gatekeeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Molten Gatekeeper — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				}),
		},
		Activated: []ActivatedAbility{Unearth("{R}")},
	})
}
