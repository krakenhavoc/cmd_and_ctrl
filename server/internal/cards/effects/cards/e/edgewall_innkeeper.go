package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edgewall Innkeeper — Creature — Human Peasant {G}, 1/1 (Throne of
// Eldraine, #5516):
//
//	"Whenever you cast a creature spell that has an Adventure, draw a card.
//	 (It doesn't need to have gone on the Adventure first.)"
//
// A creature spell is the card's permanent half, so the Adventure half never
// counts; the card has an Adventure from hand or from exile
// (CardHasAdventure).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3d7cea4-0d4c-4320-966d-17ef69b5e433",
		Name:         "Edgewall Innkeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(s82OldB142CreatureWithAdventure),
				"Edgewall Innkeeper — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

func s82OldB142CreatureWithAdventure(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && effects.CardHasAdventure(c)
}
