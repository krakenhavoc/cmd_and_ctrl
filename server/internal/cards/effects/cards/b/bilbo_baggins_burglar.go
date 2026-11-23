package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bilbo Baggins, Burglar // Take a Glance — an Adventure card (CR 715,
// The Hobbit, #4454), oracle fa85f9e6…:
//
//	Bilbo Baggins, Burglar — Legendary Creature — Halfling Rogue {2}{U}, 2/1
//	  "When Bilbo Baggins enters, draw a card."
//	Take a Glance — Sorcery — Adventure {U}
//	  "Scry 2. (Then exile this card. You may cast the creature later
//	   from exile.)"
//
// No simplification.

// bilboBagginsBurglarOracleID is shared by both faces of the card.
const BilboBagginsBurglarOracleID = "fa85f9e6-69ff-4669-b70a-440aa7b16c97"

func init() {
	effects.Register(effects.Spec{
		OracleID:     BilboBagginsBurglarOracleID,
		Name:         "Bilbo Baggins, Burglar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bilbo Baggins — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     BilboBagginsBurglarOracleID + "#1",
		Name:         "Take a Glance",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{N: 2}.Apply(ctx)
		},
	})
}
