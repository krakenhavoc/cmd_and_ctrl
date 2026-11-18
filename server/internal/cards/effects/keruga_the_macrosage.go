package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keruga, the Macrosage — Legendary Creature — Dinosaur Hippo
// {3}{G/U}{G/U}, 5/4:
//
//	"Companion — Your starting deck contains only cards with mana value 3
//	 or greater and land cards. (...)
//	 When Keruga enters, draw a card for each other permanent you control
//	 with mana value 3 or greater."
//
// Counted as the trigger resolves, over the permanents you control now
// (Keruga excluded, lands and tokens included by their mana value, which is
// 0 for a land and a token that is not a copy). Companion is the deck
// check's business (internal/deck companion registry, ADR 0172).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca5121b2-e6d9-40bc-a2d2-05852c4efbe8",
		Name:         "Keruga, the Macrosage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Keruga, the Macrosage — draw a card for each other permanent you control with mana value 3 or greater",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.InstanceID == item.SourceCardID || c.Controller != item.Controller {
							continue
						}
						if mv, ok := g.ManaValueForEffect(c); ok && mv >= 3 {
							n++
						}
					}
					return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
