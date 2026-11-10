package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "eefe2d53-4856-465d-a900-77836ef9ae07",
		Name:            "Prophet of the Scarab",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prophet of the Scarab — draw cards equal to the number of Zombies you control or Zombie cards in your graveyard, whichever is greater",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					zombie := src.TextSubtype("Zombie")
					n := cmdrB23ZombiesYouControl(g, &src, item.Controller)
					if yard := len(graveyardIDs(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(zombie) })); yard > n {
						n = yard
					}
					return DrawCards{N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
