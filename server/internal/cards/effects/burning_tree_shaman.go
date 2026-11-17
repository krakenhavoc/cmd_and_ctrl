package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burning-Tree Shaman — Creature — Centaur Shaman {1}{R}{G}, 3/4
// (Guildpact, #5599):
//
//	"Whenever a player activates an ability that isn't a mana ability,
//	 this creature deals 1 damage to that player."
//
// Harsh Mentor's watch with no source clause and every player: it
// watches EventActivateAbility, which only an activation onto the stack
// emits (CR 602.2b), and not EventManaAbilityActivated, so a mana
// ability (CR 605.1a) never triggers it. "That player" is the activator,
// read off the triggering event at resolution (triggeringActor); the
// ability does not target. The Shaman's controller is a player too, so
// their own activations trigger it. A copied ability is not activated
// and triggers nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8593e899-8e6f-4d40-96ab-98788b3d9486",
		Name:         "Burning-Tree Shaman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventActivateAbility, AnyPlayer, "Burning-Tree Shaman — 1 damage to the player who activated",
				func(g *game.Game, item *game.StackItem) error {
					activator := triggeringActor(item)
					if activator == uuid.Nil {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: activator, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
