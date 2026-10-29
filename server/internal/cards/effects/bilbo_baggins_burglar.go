package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const bilboBagginsBurglarOracleID = "fa85f9e6-69ff-4669-b70a-440aa7b16c97"

func init() {
	Register(Spec{
		OracleID:     bilboBagginsBurglarOracleID,
		Name:         "Bilbo Baggins, Burglar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bilbo Baggins — draw a card", Do(DrawCards{N: 1})),
		},
	})
	Register(Spec{
		OracleID:     bilboBagginsBurglarOracleID + "#1",
		Name:         "Take a Glance",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Scry{N: 2}.Apply(ctx)
		},
	})
}
