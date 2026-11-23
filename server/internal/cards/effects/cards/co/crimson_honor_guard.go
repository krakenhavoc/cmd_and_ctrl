package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crimson Honor Guard — Creature — Vampire Knight {3}{R}{R}, 4/5:
//
//	"Trample
//	 At the beginning of each player's end step, this creature deals 4
//	 damage to that player unless they control a commander."
//
// "That player" is the player whose end step began, the trigger event's
// actor. The commander test is made as the trigger resolves, a commander
// permanent that player controls (FreeIfYouControlCommander's reading: one
// in the command zone does not count, a stolen one does). The damage is
// dealt by the Guard even if it has left the battlefield by then
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e638aa1-31ff-4e57-9e4e-444211031eb5",
		Name:            "Crimson Honor Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Crimson Honor Guard — 4 damage to that player unless they control a commander",
				func(g *game.Game, item *game.StackItem) error {
					player := item.Trigger.Event.Actor
					if effects.ControlsACommander(g, player) {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 4}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
