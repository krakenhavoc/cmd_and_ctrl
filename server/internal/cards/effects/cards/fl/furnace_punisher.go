package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furnace Punisher — Creature — Phyrexian Warrior {2}{R}, 3/3 (Phyrexia:
// All Will Be One):
//
//	"Menace
//	 At the beginning of each player's upkeep, this creature deals 2 damage
//	 to that player unless they control two or more basic lands."
//
// The upkeep's player is read off the trigger's event; their basic lands
// are counted as the trigger resolves (CR 608.2h). Damage from a Punisher
// that has left is still dealt by its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "606af824-e711-4957-99c0-d02b0f6c729d",
		Name:            "Furnace Punisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Furnace Punisher — 2 damage to that player unless they control two or more basic lands",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					who := item.Trigger.Event.Actor
					basics := 0
					for i := range g.Battlefield.Cards {
						if c := g.Battlefield.Cards[i]; c.Controller == who && effects.IsBasicLand(c) {
							basics++
						}
					}
					if basics >= 2 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: who, Amount: 2}.Apply(ctx)
				}),
		},
	})
}
