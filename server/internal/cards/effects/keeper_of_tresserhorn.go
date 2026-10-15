package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keeper of Tresserhorn — Creature — Avatar {5}{B}, 6/6:
//
//	"Whenever this creature attacks and isn't blocked, it assigns no combat damage this turn and
//	 defending player loses 2 life."
//
// #3531: no "may", both halves happen. The defending player is the one whose block declaration
// fired the trigger.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ce976cb2-2b69-4d00-918c-d4900244ac86",
		Name:         "Keeper of Tresserhorn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlockedEffect("Keeper of Tresserhorn — it assigns no combat damage this turn and defending player loses 2 life",
				func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
					assignsNoCombatDamageThisTurn(NewContext(g, item), "Keeper of Tresserhorn — assigns no combat damage this turn")
					if !defendingPlayerStillIn(g, defender) {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -2)
				}),
		},
	})
}
