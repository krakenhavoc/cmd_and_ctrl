package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildfire Eternal — {3}{R} Creature — Zombie Jackal Cleric 1/4 (Hour of Devastation):
//
//	"Afflict 4
//	 Whenever this creature attacks and isn't blocked, you may cast an instant
//	 or sorcery spell from your hand without paying its mana cost."
//
// The cast is offered as the trigger resolves; a hand with no instant or sorcery
// asks nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "195736a9-d8b0-4c74-abe9-0289aa83c67d",
		Name:         "Wildfire Eternal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Afflict(4, "Wildfire Eternal"),
			effects.WhenAttacksAndIsNotBlockedEffect("Wildfire Eternal — you may cast an instant or sorcery spell from your hand without paying its mana cost",
				func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
					effects.OfferFreeHandCast(effects.NewContext(g, item), item.Controller, wildfireEternalFreeCast,
						"an instant or sorcery spell", effects.FreeInstantOrSorcery())
					return nil
				}),
		},
	})
}

var wildfireEternalFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/wildfire-eternal-cast", "Wildfire Eternal")
