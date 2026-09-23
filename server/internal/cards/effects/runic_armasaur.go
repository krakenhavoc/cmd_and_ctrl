package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Runic Armasaur — Creature — Dinosaur {1}{G}{G}, 2/5:
//
//	"Whenever an opponent activates an ability of a creature or land
//	 that isn't a mana ability, you may draw a card."
//
// The other half of the row #1210 closed, and the one that shows the
// shape composes: same constructor as Harsh Mentor with a narrower
// source predicate and a different payoff, wrapped in `Optional` for
// the "you may".
//
// A 2/5 for three is a body that survives a Commander table's early
// combat, and the trigger is a slow, inevitable card-advantage engine
// against exactly the decks a 2/5 cannot race: fetchlands, equipment,
// mana rocks with activated abilities, every "{T}: draw" commander.
//
// "YOU MAY DRAW" is a CR 603.4 optional trigger, so the question is
// asked as the ability would go on the stack and a "no" drops it
// without effect — the same `Optional` wrapper every other may-
// trigger uses, and the reason the payoff here is a plain DrawCards
// rather than a branch.
//
// "OF A CREATURE OR LAND" is narrower than Harsh Mentor's artifact-
// creature-land: the Armasaur does not care about mana rocks, only
// about creatures and lands. And "that isn't a mana ability" is the
// same `includeMana: false` — the trigger never watches
// EventManaAbilityActivated, so a fetchland's sacrifice triggers it
// and a Forest's {T}: Add {G} does not.
//
// The `activator` the constructor hands the payoff is unused here:
// the card draws for ITS controller, not for the player who
// activated. The argument is part of the one shape both cards share,
// and a card that ignores it says so by ignoring it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48e8ae59-a234-498e-9dae-bac8d1424ea5",
		Name:         "Runic Armasaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				WheneverAnOpponentActivates(
					"Runic Armasaur — draw a card",
					Or(Creature(), Land()),
					false,
					func(uuid.UUID) Effect { return Do(DrawCards{N: 1}) }),
				"Runic Armasaur — draw a card?"),
		},
	})
}
