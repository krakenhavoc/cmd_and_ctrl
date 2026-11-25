package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garenbrig Squire — Creature — Human Soldier {1}{G}, 2/2 (Throne of
// Eldraine, #5516):
//
//	"Whenever you cast a creature spell that has an Adventure, this
//	 creature gets +1/+1 until end of turn. (It doesn't need to have gone
//	 on the Adventure first.)"
//
// A creature spell is the card's permanent half, so the Adventure half (an
// instant or sorcery spell) never counts; the card has an Adventure whether
// cast from hand or from exile (CardHasAdventure).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a95ebaa-543e-4407-8050-adf6d8319fc1",
		Name:         "Garenbrig Squire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(s82OldB141CreatureWithAdventure),
				"Garenbrig Squire — this creature gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Garenbrig Squire — +1/+1")),
		},
	})
}

func s82OldB141CreatureWithAdventure(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && effects.CardHasAdventure(c)
}
