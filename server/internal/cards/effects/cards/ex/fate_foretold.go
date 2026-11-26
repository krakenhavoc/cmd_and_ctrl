package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fate Foretold — Enchantment — Aura {1}{U} (Theros, S82 OLD b390, #5561):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 When enchanted creature dies, its controller draws a card."
//
// Spreading Seas' enters trigger beside Minion's Return's dies trigger. "Its
// controller" is the creature's controller as it last existed on the
// battlefield (CR 603.10a, CR 608.2h): a stolen creature draws for the thief,
// not the owner.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d578ef54-6a6b-4839-a059-261728fdaef1",
		Name:         "Fate Foretold",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fate Foretold — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.WhenEnchantedCreatureDies("Fate Foretold — its controller draws a card",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					who := ev.LastKnownController
					if who == uuid.Nil {
						if c, ok := g.LookupCardForEffect(ev.CardID); ok {
							who = c.Controller
						}
					}
					return effects.DrawCards{Player: who, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
