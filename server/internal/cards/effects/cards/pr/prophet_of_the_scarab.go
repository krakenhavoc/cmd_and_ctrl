package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prophet of the Scarab — Creature — Zombie Wizard {4}{U}, 3/4:
//
//	"Vigilance
//	 When this creature enters, draw cards equal to the number of Zombies you
//	 control or the number of Zombie cards in your graveyard, whichever is
//	 greater."
//
// Both counts are made as the trigger resolves, the Prophet itself among the
// Zombies you control if it is still there. The graveyard count reads the
// cards' printed subtypes (CR 109.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eefe2d53-4856-465d-a900-77836ef9ae07",
		Name:            "Prophet of the Scarab",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Prophet of the Scarab — draw cards equal to the number of Zombies you control or Zombie cards in your graveyard, whichever is greater",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					zombie := src.TextSubtype("Zombie")
					n := effects.CmdrB23ZombiesYouControl(g, &src, item.Controller)
					if yard := len(effects.GraveyardIDs(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(zombie) })); yard > n {
						n = yard
					}
					return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
